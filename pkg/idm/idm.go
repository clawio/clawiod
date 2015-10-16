package idm

const (
	MockIDM = iota
	FileIDM
)

type Type int

type IDM struct {
}

func (i *IDM) CreateIDM(t Type) {

}
