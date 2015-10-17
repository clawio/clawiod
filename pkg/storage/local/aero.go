package local

import (
	as "github.com/aerospike/aerospike-client-go"
	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/storage"
	"strings"
	"time"
)

// Aero is the the distributed KV store to keep track of mtime and etag
// propagation.
// It also keeps a map of <resourceID> => <resourcePath> to make operations
// that depend on ID faster. Sharing will use id-based operations.
type Aero struct {
	c   *as.Client
	cfg config.Config
}

// NewAero returns a new Aero client.
func NewAero(cfg config.Config) (*Aero, error) {
	c, err := as.NewClient(cfg.GetDirectives().LocalStorageAeroSpikeHost,
		cfg.GetDirectives().LocalStorageAeroSpikePort)

	if err != nil {
		return nil, err
	}

	return &Aero{c: c, cfg: cfg}, nil
}

func (a *Aero) PutRecord(resourcePath, resourceID string) error {
	return a.putRecord(resourcePath, resourceID)
}

func (a *Aero) GetRecord(resourcePath string) (*as.Record, error) {
	return a.getRecord(resourcePath)
}
func (a *Aero) GetOrCreateRecord(resourcePath string) (*as.Record, error) {
	r, err := a.getRecord(resourcePath)
	if err != nil {
		return nil, err
	}
	if r == nil { // not found
		// create it
		err := a.PutRecord(resourcePath, uuid.New())
		if err != nil {
			return nil, err
		}

		r, err = a.GetRecord(resourcePath)
		if err != nil {
			return nil, err
		}

		if r == nil {
			// it is removed after the storage commit and the aero insert.
			return nil, &storage.NotExistError{}
		}
	}
	return r, nil
}
func (a *Aero) putRecord(resourcePath, resourceID string) error {
	err := a.propagateChanges(resourcePath)
	if err != nil {
		return err
	}
	return a.insertID(resourcePath, resourceID)
}

func (a *Aero) getRecord(resourcePath string) (*as.Record, error) {
	k, err := as.NewKey(a.cfg.GetDirectives().LocalStorageAeroSpikeNamespace,
		a.cfg.GetDirectives().LocalStorageAeroSpikePropagatorSet, resourcePath)

	r, err := a.c.Get(nil, k)
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (a *Aero) insertID(resourcePath, resourceID string) error {
	bins := []*as.Bin{
		as.NewBin("ts", time.Now().UnixNano()),
		as.NewBin("path", resourcePath),
	}

	k, err := as.NewKey(a.cfg.GetDirectives().LocalStorageAeroSpikeNamespace,
		a.cfg.GetDirectives().LocalStorageAeroSpikeRID2PathSet, resourceID)

	if err != nil {
		return err
	}

	pol := as.NewWritePolicy(0, -1)
	pol.SendKey = true

	err = a.c.PutBins(pol, k, bins...)
	if err != nil {
		return err
	}

	return nil

}
func (a *Aero) propagateChanges(resourcePath string) error {

	bins := []*as.Bin{
		as.NewBin("mtime", time.Now().UnixNano()),
		as.NewBin("etag", uuid.New()),
	}

	err := a.updateModification(resourcePath, bins)
	if err != nil {
		return err
	}

	// Propagate mtime and etag bottom up, just until the identity PID.
	parents := strings.Split(resourcePath, "/")
	parents = parents[:len(parents)-1] // remove actual object

	for len(parents) > 0 {
		p := strings.Join(parents, "/")

		// go one level up independently of errors
		parents = parents[:len(parents)-1]

		err := a.updateModification(p, bins)

		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Aero) updateModification(resourcePath string, bins []*as.Bin) error {

	k, err := as.NewKey(a.cfg.GetDirectives().LocalStorageAeroSpikeNamespace,
		a.cfg.GetDirectives().LocalStorageAeroSpikePropagatorSet, resourcePath)

	if err != nil {
		return err
	}

	pol := as.NewWritePolicy(0, -1)
	pol.SendKey = true

	err = a.c.PutBins(pol, k, bins...)
	if err != nil {
		return err
	}

	return nil
}

/*
import (
	"github.com/clawio/clawiod/pkg/config"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"path"
	"strings"
)

type CatalogEntry struct {
	ID        int    `gorm:"primary_key", sql:"AUTO_INCREMENT"`
	Path      string `sql:"size:255"`
	Mtime     int
	Etag      string
	Container bool
}

type MetaStore struct {
	*gorm.DB
}

func NewMetaStore(cfg config.Config) (*MetaStore, error) {
	// db, err := gorm.Open("postgres", "user=gorm dbname=gorm sslmode=disable")
	// db, err := gorm.Open("foundation", "dbname=gorm") // FoundationDB.
	// db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	db, err := gorm.Open("sqlite3", cfg.GetDirectives().LocalStorageMetaStoreSQLite3)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	db.SingularTable(true)

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(1000)

	res := db.Set("gorm:table_options", "ENGINE=InnoDB")
	if res.Error != nil {
		return nil, res.Error
	}

	res = db.AutoMigrate(&CatalogEntry{})
	if res.Error != nil {
		return nil, res.Error
	}
	return &MetaStore{DB: &db}, nil
}

func (m *MetaStore) GetEntryByID(id string) (*CatalogEntry, error) {
	entry := &CatalogEntry{}
	res := m.First(entry, id)
	if res.Error != nil {
		return nil, res.Error
	}
	return entry, nil
}

func (m *MetaStore) Insert(e *CatalogEntry) error {
	res := m.Create(e)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (m *MetaStore) GetEntryByPath(resourcePath string) (*CatalogEntry, bool, error) {
	entry := &CatalogEntry{}
	res := m.Where("path = ?", resourcePath).First(entry)
	if res.Error != nil && !res.RecordNotFound() {
		return nil, false, res.Error
	}
	if res.RecordNotFound() {
		return nil, false, nil
	}
	return entry, true, nil
}

func (m *MetaStore) Update(e *CatalogEntry) (bool, error) {
	res := m.Save(e)
	if res.Error != nil && !res.RecordNotFound() {
		return false, res.Error
	}
	if res.RecordNotFound() {
		return false, nil
	}
	return true, nil
}

// Upsert inserts or updates an existing entry and propagates changes to parent folders in an atomic way.
func (m *MetaStore) Upsert(e *CatalogEntry) (*CatalogEntry, error) {
	newMtime := e.Mtime
	newEtag := e.Etag

	tmp := &CatalogEntry{}

	tx := m.Begin()
	// Check if the resource with that path already exists
	res := tx.Where("path=?", e.Path).First(tmp)
	if res.Error != nil && !res.RecordNotFound() {
		tx.Rollback()
		return nil, res.Error
	}

	// Create new entry
	if res.RecordNotFound() {
		res := tx.Create(e)
		if res.Error != nil {
			tx.Rollback()
			return nil, res.Error
		}
	} else {
		// Update existing entry
		e.ID = tmp.ID // give the received entry the ID of the already existing.
		res = tx.Save(e)
		if res.Error != nil {
			tx.Rollback()
			return nil, res.Error
		}
	}

	// Propagate mtime and etag bottom up, just until the identity PID.
	parents := strings.Split(e.Path, "/")
	parents = parents[:len(parents)-1] // remove actual object
	for len(parents) > 0 {
		tmp := &CatalogEntry{}
		p := path.Join(parents...) + "/" // containers end with final slash.

		// Check if parent exists
		res := tx.Where("path=?", p).First(tmp)
		if res.Error != nil && !res.RecordNotFound() {
			tx.Rollback()
			return nil, res.Error
		}

		// Create parent folder
		if res.RecordNotFound() {
			tmp.Mtime = newMtime
			tmp.Etag = newEtag
			tmp.Container = true
			tmp.Path = p
			res := tx.Create(tmp)
			if res.Error != nil {
				tx.Rollback()
				return nil, res.Error
			}
		} else {
			// Update existing parent
			tmp.Mtime = newMtime
			tmp.Etag = newEtag
			res := tx.Model(tmp).Where("path=?", p).Save(tmp)
			if res.Error != nil {
				tx.Rollback()
				return nil, res.Error
			}
		}

		// go one level up
		parents = parents[:len(parents)-1]
	}

	// Finish transaction
	res = tx.Commit()
	if res.Error != nil {
		return nil, res.Error
	}
	return e, nil
}
*/
