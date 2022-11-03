package types

type Type interface {
	isType()
}

// type BasicKind int
//
// const (
// 	Invalid BasicKind = iota
// 	Bool
// 	String
// 	Integer
// 	Float
// )
//
// type Basic struct {
// 	Kind BasicKind
// 	Name string
// }

type WellType struct {
	Name string
	// Value interface{}
}

// func (Basic) isType() {}
func (WellType) isType() {}
