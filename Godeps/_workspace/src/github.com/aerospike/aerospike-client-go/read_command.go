// Copyright 2013-2015 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"math"
	"reflect"
	"strings"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"

	. "github.com/aerospike/aerospike-client-go/types"
	Buffer "github.com/aerospike/aerospike-client-go/utils/buffer"
)

type readCommand struct {
	*singleCommand

	policy   Policy
	binNames []string
	record   *Record

	// pointer to the object that's going to be unmarshalled
	object interface{}
}

func newReadCommand(cluster *Cluster, policy Policy, key *Key, binNames []string) *readCommand {
	return &readCommand{
		singleCommand: newSingleCommand(cluster, key),
		binNames:      binNames,
		policy:        policy,
	}
}

func (cmd *readCommand) getPolicy(ifc command) Policy {
	return cmd.policy
}

func (cmd *readCommand) writeBuffer(ifc command) error {
	return cmd.setRead(cmd.policy.GetBasePolicy(), cmd.key, cmd.binNames)
}

func (cmd *readCommand) parseResult(ifc command, conn *Connection) error {
	// Read header.
	_, err := conn.Read(cmd.dataBuffer, int(_MSG_TOTAL_HEADER_SIZE))
	if err != nil {
		Logger.Warn("parse result error: " + err.Error())
		return err
	}

	// A number of these are commented out because we just don't care enough to read
	// that section of the header. If we do care, uncomment and check!
	sz := Buffer.BytesToInt64(cmd.dataBuffer, 0)
	headerLength := int(cmd.dataBuffer[8])
	resultCode := ResultCode(cmd.dataBuffer[13] & 0xFF)
	generation := int(Buffer.BytesToUint32(cmd.dataBuffer, 14))
	expiration := TTL(int(Buffer.BytesToUint32(cmd.dataBuffer, 18)))
	fieldCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 26)) // almost certainly 0
	opCount := int(Buffer.BytesToUint16(cmd.dataBuffer, 28))
	receiveSize := int((sz & 0xFFFFFFFFFFFF) - int64(headerLength))

	// Read remaining message bytes.
	if receiveSize > 0 {
		if err = cmd.sizeBufferSz(receiveSize); err != nil {
			return err
		}
		_, err = conn.Read(cmd.dataBuffer, receiveSize)
		if err != nil {
			Logger.Warn("parse result error: " + err.Error())
			return err
		}

	}

	if resultCode != 0 {
		if resultCode == KEY_NOT_FOUND_ERROR && cmd.object == nil {
			return nil
		}

		if resultCode == UDF_BAD_RESPONSE {
			cmd.record, _ = cmd.parseRecord(opCount, fieldCount, generation, expiration)
			err := cmd.handleUdfError(resultCode)
			Logger.Warn("UDF execution error: " + err.Error())
			return err
		}

		return NewAerospikeError(resultCode)
	}

	if cmd.object == nil {
		if opCount == 0 {
			// data Bin was not returned.
			cmd.record = newRecord(cmd.node, cmd.key, nil, generation, expiration)
			return nil
		}

		cmd.record, err = cmd.parseRecord(opCount, fieldCount, generation, expiration)
		if err != nil {
			return err
		}
	} else {
		cmd.parseObject(opCount, fieldCount, generation, expiration)
	}

	return nil
}

func (cmd *readCommand) handleUdfError(resultCode ResultCode) error {
	if ret, exists := cmd.record.Bins["FAILURE"]; exists {
		return NewAerospikeError(resultCode, ret.(string))
	}
	return NewAerospikeError(resultCode)
}

func (cmd *readCommand) parseRecord(
	opCount int,
	fieldCount int,
	generation int,
	expiration int,
) (*Record, error) {
	var bins BinMap
	receiveOffset := 0

	// There can be fields in the response (setname etc).
	// But for now, ignore them. Expose them to the API if needed in the future.
	// Logger.Debug("field count: %d, databuffer: %v", fieldCount, cmd.dataBuffer)
	if fieldCount > 0 {
		// Just skip over all the fields
		for i := 0; i < fieldCount; i++ {
			// Logger.Debug("%d", receiveOffset)
			fieldSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
			receiveOffset += (4 + fieldSize)
		}
	}

	for i := 0; i < opCount; i++ {
		opSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
		particleType := int(cmd.dataBuffer[receiveOffset+5])
		nameSize := int(cmd.dataBuffer[receiveOffset+7])
		name := string(cmd.dataBuffer[receiveOffset+8 : receiveOffset+8+nameSize])
		receiveOffset += 4 + 4 + nameSize

		particleBytesSize := int(opSize - (4 + nameSize))
		value, _ := bytesToParticle(particleType, cmd.dataBuffer, receiveOffset, particleBytesSize)
		receiveOffset += particleBytesSize

		if bins == nil {
			bins = make(BinMap, opCount)
		}
		bins[name] = value
	}

	return newRecord(cmd.node, cmd.key, bins, generation, expiration), nil
}

func (cmd *readCommand) parseObject(
	opCount int,
	fieldCount int,
	generation int,
	expiration int,
) error {
	receiveOffset := 0

	// There can be fields in the response (setname etc).
	// But for now, ignore them. Expose them to the API if needed in the future.
	// Logger.Debug("field count: %d, databuffer: %v", fieldCount, cmd.dataBuffer)
	if fieldCount > 0 {
		// Just skip over all the fields
		for i := 0; i < fieldCount; i++ {
			// Logger.Debug("%d", receiveOffset)
			fieldSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
			receiveOffset += (4 + fieldSize)
		}
	}

	var rv reflect.Value
	if opCount > 0 {
		rv = reflect.ValueOf(cmd.object).Elem()

		// map tags
		cacheObjectTags(rv)
	}

	for i := 0; i < opCount; i++ {
		opSize := int(Buffer.BytesToUint32(cmd.dataBuffer, receiveOffset))
		particleType := int(cmd.dataBuffer[receiveOffset+5])
		nameSize := int(cmd.dataBuffer[receiveOffset+7])
		name := string(cmd.dataBuffer[receiveOffset+8 : receiveOffset+8+nameSize])
		receiveOffset += 4 + 4 + nameSize

		particleBytesSize := int(opSize - (4 + nameSize))
		value, _ := bytesToParticle(particleType, cmd.dataBuffer, receiveOffset, particleBytesSize)
		if err := cmd.setObjectField(rv, name, value); err != nil {
			return err
		}

		receiveOffset += particleBytesSize
	}

	return nil
}

func (cmd *readCommand) GetRecord() *Record {
	return cmd.record
}

func (cmd *readCommand) Execute() error {
	return cmd.execute(cmd)
}

func (cmd *readCommand) setObjectField(obj reflect.Value, fieldName string, value interface{}) error {
	if value == nil {
		return nil
	}

	// find the name based on tag mapping
	iobj := reflect.Indirect(obj)
	if name, exists := objectMappings.getMapping(iobj)[fieldName]; exists {
		fieldName = name
	}
	f := iobj.FieldByName(fieldName)
	setValue(f, value)

	return nil
}

func setValue(f reflect.Value, value interface{}) error {
	// find the name based on tag mapping
	if f.CanSet() {
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(int64(value.(int)))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch v := value.(type) {
			case uint8:
				f.SetUint(uint64(v))
			case int:
				f.SetUint(uint64(v))
			default:
				f.SetUint(value.(uint64))
			}
		case reflect.Float64, reflect.Float32:
			f.SetFloat(float64(math.Float64frombits(uint64(value.(int)))))
		case reflect.String:
			rv := reflect.ValueOf(value.(string))
			if rv.Type() != f.Type() {
				rv = rv.Convert(f.Type())
			}
			f.Set(rv)
		case reflect.Bool:
			f.SetBool(value.(int) == 1)
		case reflect.Interface:
			if value != nil {
				f.Set(reflect.ValueOf(value))
			}
		case reflect.Ptr:
			switch f.Type().Elem().Kind() {
			case reflect.Int:
				tempV := int(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint:
				tempV := uint(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.String:
				tempV := string(value.(string))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int8:
				tempV := int8(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint8:
				tempV := uint8(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int16:
				tempV := int16(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint16:
				tempV := uint16(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int32:
				tempV := int32(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint32:
				tempV := uint32(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Int64:
				tempV := int64(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Uint64:
				tempV := uint64(value.(int))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Float64:
				tempV := math.Float64frombits(uint64(value.(int)))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Bool:
				tempV := bool(value.(int) == 1)
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Float32:
				tempV := float32(math.Float64frombits(uint64(value.(int))))
				rv := reflect.ValueOf(&tempV)
				if rv.Type() != f.Type() {
					rv = rv.Convert(f.Type())
				}
				f.Set(rv)
			case reflect.Interface:
				f.Set(reflect.ValueOf(&value))
			case reflect.Struct:
				// support time.Time
				if f.Type().Elem().PkgPath() == "time" && f.Type().Elem().Name() == "Time" {
					tm := time.Unix(0, int64(value.(int)))
					f.Set(reflect.ValueOf(&tm))
					break
				} else {
					valMap := value.(map[interface{}]interface{})
					// iteraste over struct fields and recursively fill them up
					if valMap != nil {
						newObjPtr := f
						if f.IsNil() {
							newObjPtr = reflect.New(f.Type().Elem())
						}
						theStruct := newObjPtr.Elem().Type()
						numFields := newObjPtr.Elem().NumField()
						for i := 0; i < numFields; i++ {
							// skip unexported fields
							if theStruct.Field(i).PkgPath != "" {
								continue
							}

							alias := theStruct.Field(i).Name
							tag := strings.Trim(theStruct.Field(i).Tag.Get(aerospikeTag), " ")
							if tag != "" {
								alias = tag
							}

							if valMap[alias] != nil {
								setValue(reflect.Indirect(newObjPtr).FieldByName(alias), valMap[alias])
							}
						}

						// set the field
						f.Set(newObjPtr)
					}
				}
			} // switch ptr
		case reflect.Slice, reflect.Array:
			// BLOBs come back as []byte
			theArray := reflect.ValueOf(value)

			if f.Kind() == reflect.Slice {
				if f.IsNil() {
					f.Set(reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), theArray.Len(), theArray.Len()))
				} else if f.Len() < theArray.Len() {
					count := theArray.Len() - f.Len()
					f = reflect.AppendSlice(f, reflect.MakeSlice(reflect.SliceOf(f.Type().Elem()), count, count))
				}
			}

			for i := 0; i < theArray.Len(); i++ {
				setValue(f.Index(i), theArray.Index(i).Interface())
			}
		case reflect.Map:
			theMap := value.(map[interface{}]interface{})
			if theMap != nil {
				newMap := reflect.MakeMap(f.Type())
				var newKey, newVal reflect.Value
				for key, elem := range theMap {
					if key != nil {
						newKey = reflect.ValueOf(key)
					} else {
						newKey = reflect.Zero(f.Type().Key())
					}

					if elem != nil {
						newVal = reflect.ValueOf(elem)
					} else {
						newVal = reflect.Zero(f.Type().Elem())
					}

					newMap.SetMapIndex(newKey, newVal)
				}
				f.Set(newMap)
			}

		case reflect.Struct:
			// support time.Time
			if f.Type().PkgPath() == "time" && f.Type().Name() == "Time" {
				f.Set(reflect.ValueOf(time.Unix(0, int64(value.(int)))))
				break
			}

			valMap := value.(map[interface{}]interface{})
			// iteraste over struct fields and recursively fill them up
			typeOfT := f.Type()
			numFields := f.NumField()
			for i := 0; i < numFields; i++ {
				// skip unexported fields
				if typeOfT.Field(i).PkgPath != "" {
					continue
				}

				alias := typeOfT.Field(i).Name
				tag := strings.Trim(typeOfT.Field(i).Tag.Get(aerospikeTag), " ")
				if tag != "" {
					alias = tag
				}

				if valMap[alias] != nil {
					setValue(f.FieldByName(typeOfT.Field(i).Name), valMap[alias])
				}
			}

			// set the field
			f.Set(f)
		}
	}

	return nil
}
