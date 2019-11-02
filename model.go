package gava

type ClassField struct {
	TypeCode  byte
	Name      string
	className string
	Value     string
}

type ClassDetails struct {
	ClassName        string
	RefHandle        int
	ClassDescFlags   byte
	FieldDescription []*ClassField
	ObjectValue string
}

type ClassDataDesc struct {
	ClassDetail []*ClassDetails
}

func (cd *ClassDataDesc) buildClassDataDescFromIndex(index int) *ClassDataDesc {
	list := []*ClassDetails{}
	for i := index; i < len(cd.ClassDetail); i++ {
		list = append(list, cd.ClassDetail[i])
	}
	return &ClassDataDesc{ClassDetail: list}
}
