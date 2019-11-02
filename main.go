package gava

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)

type GavaDeserilizer struct {
	handleValue           int
	classDataDescriptions []*ClassDataDesc
	data                  []byte
}

func NewGavaDeserilizer(data []byte) *GavaDeserilizer {
	return &GavaDeserilizer{
		handleValue:           0x7e0000,
		classDataDescriptions: []*ClassDataDesc{},
		data:                  data,
	}
}

func (g *GavaDeserilizer) Parse() *ClassDetails {
	var b1 byte
	var b2 byte

	//The stream may begin with an RMI packet type byte, print it if so
	if g.data[0] != 0xac {
		b1 = g.data[0]
		g.data = g.data[1:]

		switch b1 {
		case 0x50:
			//fmt.Println("RMI Call - 0x50")
			break
		case 0x51:
			//fmt.Println("RMI ReturnData - 0x51")
			break
		case 0x52:
			//fmt.Println("RMI Ping - 0x52")
			break
		case 0x53:
			//fmt.Println("RMI PingAck - 0x53")
			break
		case 0x54:
			//fmt.Println("RMI DgcAck - 0x54")
			break
		default:
			//fmt.Println("Unknown RMI packet type - 0x" + hex.EncodeToString([]byte{b1}))
			break
		}
	}

	//Magic number, print and validate
	b1 = g.data[0]
	g.data = g.data[1:]
	b2 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("STREAM_MAGIC - 0x" + hex.EncodeToString([]byte{b1}) + " " + hex.EncodeToString([]byte{b2}))
	if b1 != 0xac || b2 != 0xed {
		//fmt.Println("Invalid STREAM_MAGIC, should be 0xac ed")
		return nil
	}

	//Serialization version
	b1 = g.data[0]
	g.data = g.data[1:]
	b2 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("STREAM_VERSION - 0x" + hex.EncodeToString([]byte{b1}) + " " + hex.EncodeToString([]byte{b2}))
	if b1 != 0x00 || b2 != 0x05 {
		//fmt.Println("Invalid STREAM_VERSION, should be 0x00 05")
	}

	//fmt.Println("Contents")
	for len(g.data) > 0 {
		c, _ := g.readContentElement()
		return c
	}

	return nil
}

func (g *GavaDeserilizer) readContentElement() (*ClassDetails, string) {
	switch g.data[0] {
	case 0x73: //TC_OBJECT
		return g.readNewObject(), ""
	case 0x76: //TC_CLASS
		g.readNewClass()
		break
	case 0x75: //TC_ARRAY
		return nil, g.readNewArray()
	case 0x74: //TC_STRING
		fallthrough
	case 0x7c: //TC_LONGSTRING
		return nil, g.readNewString()
	case 0x7e: //TC_ENUM
		return nil, g.readNewEnum()
	case 0x72: //TC_CLASSDESC
		fallthrough
	case 0x7d: //TC_PROXYCLASSDESC
		g.readNewClassDesc()
		break
	case 0x71: //TC_REFERENCE
		g.readPrevObject()
		break
	case 0x70: //TC_NULL
		g.readNullReference()
		break
		//			case 0x7b:		//TC_EXCEPTION
		//				readException()
		//				break
		//			case 0x79:		//TC_RESET
		//				handleReset()
		//				break
	case 0x77: //TC_BLOCKDATA
		return nil, g.readBlockData()
	case 0x7a: //TC_BLOCKDATALONG
		return nil, g.readLongBlockData()
	default:
		print("Invalid content element type 0x" + hex.EncodeToString([]byte{g.data[0]}))
		log.Fatal("Error: Illegal content element type.")
	}
	return nil, ""
}

func (g *GavaDeserilizer) readBlockData() string {
	var b1 = g.data[0]
	g.data = g.data[1:]
	//fmt.Println("TC_BLOCK_DATA - 0x", hex.EncodeToString([]byte{b1}))
	if b1 != 0x77 {
		log.Fatal("b1 != 0x77")
	}
	var len = g.data[0] & 0xFF
	g.data = g.data[1:]

	//fmt.Println(fmt.Sprintf("Length - %d", len))

	var contents = ""

	for i := byte(0); i < len; i++ {
		contents += hex.EncodeToString([]byte{g.data[0]})
		g.data = g.data[1:]
	}
	//fmt.Println(fmt.Sprintf("Contents - 0x%s", contents))
	return contents
}

func (g *GavaDeserilizer) readLongBlockData() string {
	var b1 = g.data[0]
	g.data = g.data[1:]
	//fmt.Println("TC_BLOCK_DATA_LONG - 0x", hex.EncodeToString([]byte{b1}))
	if b1 != 0x7a {
		log.Fatal("b1 != 0x7a")
	}

	var len = int(binary.BigEndian.Uint32(g.data[0:4]))
	g.data = g.data[4:]
	//fmt.Println(fmt.Sprintf("Length - %d", len))

	var contents = ""

	for i := 0; i < len; i++ {
		contents += hex.EncodeToString([]byte{g.data[0]})
		g.data = g.data[1:]
	}

	//fmt.Println(fmt.Sprintf("Contents - 0x%s", contents))
	return contents
}

func (g *GavaDeserilizer) readNullReference() string {
	var b1 = g.data[0]
	g.data = g.data[1:]
	//fmt.Println("TC_NULL - 0x" + hex.EncodeToString([]byte{b1}))
	if b1 != 0x70 {
		log.Fatal("Error: Illegal value for TC_NULL (should be 0x70)")
	}
	return "null"
}

func (g *GavaDeserilizer) readPrevObject() int {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_REFERENCE - 0x" + hex.EncodeToString([]byte{b1}))

	if b1 != 0x71 {
		log.Fatal("b1 != 0x71")
	}

	handle := int(binary.BigEndian.Uint32(g.data[0:4]))
	g.data = g.data[4:]
	//fmt.Println(fmt.Sprintf("Handle - %d", handle))
	return handle
}

func (g *GavaDeserilizer) readNewClassDesc() *ClassDataDesc {
	switch g.data[0] {
	case 0x72:
		cdd := g.readTCClassDesc()
		g.classDataDescriptions = append(g.classDataDescriptions, cdd)
		return cdd
	case 0x7d:
		return g.readTCProxyClassDesc()
	default:
		// print("Invalid newClassDesc type 0x" + this.byteToHex(this._data.peek()));
		log.Fatal("Error illegal newClassDesc type.")
	}
	return nil
}

func (g *GavaDeserilizer) readTCProxyClassDesc() *ClassDataDesc {
	return nil
}

func (g *GavaDeserilizer) readClassDescInfo(cdd *ClassDataDesc) {
	var classDescFlags string
	var b1 = g.data[0]
	g.data = g.data[1:]

	if (b1 & 0x01) == 0x01 {
		classDescFlags += "SC_WRITE_METHOD | "
	}
	if (b1 & 0x02) == 0x02 {
		classDescFlags += "SC_SERIALIZABLE | "
	}
	if (b1 & 0x04) == 0x04 {
		classDescFlags += "SC_EXTERNALIZABLE | "
	}
	if (b1 & 0x08) == 0x08 {
		classDescFlags += "SC_BLOCKDATA | "
	}

	if classDescFlags != "" {
		classDescFlags = classDescFlags[:len(classDescFlags)-3]
	}
	//fmt.Println("classDescFlags - 0x" + hex.EncodeToString([]byte{b1}) + " - " + classDescFlags)

	//Store the classDescFlags
	cdd.ClassDetail[len(cdd.ClassDetail)-1].ClassDescFlags = b1 //Set the classDescFlags for the most recently added class

	//Validate classDescFlags
	if (b1 & 0x02) == 0x02 {
		if (b1 & 0x04) == 0x04 { log.Fatal("Error: Illegal classDescFlags, SC_SERIALIZABLE is not compatible with SC_EXTERNALIZABLE."); }
		if (b1 & 0x08) == 0x08 { log.Fatal("Error: Illegal classDescFlags, SC_SERIALIZABLE is not compatible with SC_BLOCKDATA."); }
	} else if (b1 & 0x04) == 0x04 {
		if (b1 & 0x01) == 0x01 { log.Fatal("Error: Illegal classDescFlags, SC_EXTERNALIZABLE is not compatible with SC_WRITE_METHOD."); }
	} else if b1 != 0x00 {
		log.Fatal("Error: Illegal classDescFlags, must include either SC_SERIALIZABLE or SC_EXTERNALIZABLE.");
	}
	//
	//fields
	g.readFields(cdd) //Read field descriptions and add them to the ClassDataDesc
	//
	//classAnnotation
	g.readClassAnnotation()
	//
	//superClassDesc
	scdd := g.readSuperClassDesc() //Read the super class description and add it to the ClassDataDesc
	if scdd != nil {
		for i := 0; i < len(scdd.ClassDetail); i++ {
			cdd.ClassDetail = append(cdd.ClassDetail, scdd.ClassDetail[i])
		}
	}
}

func (g *GavaDeserilizer) readClassAnnotation() {
	//fmt.Println("classAnnotations")
	for g.data[0] != 0x78 {
		g.readContentElement()
	}
	g.data = g.data[1:]
	//fmt.Println("TC_END_BLOCK_DATA - 0x78")
}

func (g *GavaDeserilizer) readSuperClassDesc() *ClassDataDesc {
	//fmt.Println("superClassDesc")
	cdd := g.readClassDesc()
	return cdd
}

func (g *GavaDeserilizer) readFields(cdd *ClassDataDesc) {
	var b1 byte
	var b2 byte
	var count uint16

	b1 = g.data[0]
	g.data = g.data[1:]
	b2 = g.data[0]
	g.data = g.data[1:]

	//numBytes := []byte{b1, b2}
	//count = binary.BigEndian.Uint16(numBytes)
	//count = (int(b1 << 8) & 0xff00) + int(b2 & 0xff)
	count = uint16(b2) | uint16(b1)<<8

	//fmt.Println(fmt.Sprintf("fieldCount - %d - 0x"+hex.EncodeToString([]byte{b1})+" "+hex.EncodeToString([]byte{b2}), count))

	if count > 0 {
		//fmt.Println("Fields")

		for i := 0; i < int(count); i++ {
			//fmt.Println(fmt.Sprintf("%d : ", i))
			g.readFieldDesc(cdd)
		}
	}
}

func (g *GavaDeserilizer) readFieldDesc(cdd *ClassDataDesc) {
	var b1 = g.data[0]
	g.data = g.data[1:]

	field := &ClassField{TypeCode: b1}
	cdd.ClassDetail[len(cdd.ClassDetail)-1].FieldDescription = append(cdd.ClassDetail[len(cdd.ClassDetail)-1].FieldDescription, field)

	switch b1 {
	case 'B':
		//fmt.Println("Byte")
	case 'C':
		//fmt.Println("Char")
	case 'D':
		//fmt.Println("Double")
	case 'F':
		//fmt.Println("Float")
	case 'I':
		//fmt.Println("Int")
	case 'J':
		//fmt.Println("Long")
	case 'S':
		//fmt.Println("Short")
	case 'Z':
		//fmt.Println("Boolean")
	case '[':
		//fmt.Println("Array")
	case 'L':
		//fmt.Println("Object")
	default:
		log.Fatal("Error: Illegal field type code ('" + string(b1) + "', 0x" + hex.EncodeToString([]byte{b1}) + ")")
	}

	//fmt.Println("fieldName")

	fieldName := g.readUtf()
	field.Name = fieldName

	if b1 == '[' || b1 == 'L' {
		//fmt.Println("className1")
		field.className = g.readNewString()
	}
}

func (g *GavaDeserilizer) readUtf() string {
	content := ""
	hexStr := ""
	var b1 byte
	var b2 byte
	var len int

	//length
	b1 = g.data[0]
	g.data = g.data[1:]
	b2 = g.data[0]
	g.data = g.data[1:]
	len = (int(b1<<8) & 0xff00) + int(b2&0xff)

	//fmt.Println("Length - " + string(len) + " - 0x" + hex.EncodeToString([]byte{b1}) + " " + hex.EncodeToString([]byte{b2}))

	//Contents
	for i := 0; i < len; i++ {
		b1 = g.data[0]
		g.data = g.data[1:]
		content += string(b1)
		hexStr += hex.EncodeToString([]byte{b1})
	}
	//fmt.Println("Value - " + content + " - 0x" + hexStr)

	return content
}

func (g *GavaDeserilizer) readTCClassDesc() *ClassDataDesc {
	var cdd = &ClassDataDesc{}
	// var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_CLASSDESC - 0x" + hex.EncodeToString([]byte{b1}))
	// if(b1 != (byte)0x72) { throw new RuntimeException("Error: Illegal value for TC_CLASSDESC (should be 0x72)"); }
	//fmt.Println("className")

	className := g.readUtf()
	cdd.ClassDetail = append(cdd.ClassDetail, &ClassDetails{
		ClassName: className,
	})

	//this.print("serialVersionUID - 0x" + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()) +
	//				   " " + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()) + " " + this.byteToHex(this._data.pop()));
	_ = hex.EncodeToString(g.data[0:8])
	g.data = g.data[8:]

	cdd.ClassDetail[0].RefHandle = g.handleValue
	g.handleValue++

	g.readClassDescInfo(cdd)

	return cdd
}

func (g *GavaDeserilizer) readNewEnum() string {
	return ""
}

func (g *GavaDeserilizer) readNewString() string {
	switch g.data[0] {
	case 0x74:
		return g.readTCString()
	case 0x7c:
		return g.readTCLongString()
	case 0x71:
		g.readPrevObject()
		return "[TC_REF]"
	default:
		log.Fatal("Error illegal newString type.")
	}
	return ""
}

func (g *GavaDeserilizer) readTCString() string {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_STRING - 0x" + hex.EncodeToString([]byte{b1}))

	if b1 != 0x74 {
		log.Fatal("Error: Illegal value for TC_STRING (should be 0x74)")
	}

	g.handleValue++

	return g.readUtf()
}

func (g *GavaDeserilizer) readTCLongString() string {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_LONG_STRING - 0x" + hex.EncodeToString([]byte{b1}))

	if b1 != 0x74 {
		log.Fatal("Error: Illegal value for TC_STRING (should be 0x74)")
	}

	g.handleValue++

	return g.readLongUtf()
}

func (g *GavaDeserilizer) readLongUtf() string {
	var content string
	var hexStr string

	length := int64(binary.BigEndian.Uint64(g.data[0:9]))
	g.data = g.data[8:]

	//fmt.Println(fmt.Sprintf("Length - %d", length))
	for i := int64(0); i < length; i++ {
		var b1 = g.data[0]
		g.data = g.data[1:]
		content += string(b1)
		hexStr += hex.EncodeToString([]byte{b1})
	}

	//fmt.Println(fmt.Sprintf("Value - %s - 0x%s", content, hexStr))

	return content
}

func (g *GavaDeserilizer) readNewArray() string {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_ARRAY - 0x" + hex.EncodeToString([]byte{b1}))

	if b1 != 0x75 {
		log.Fatal("b1 != 0x75")
	}

	cdd := g.readClassDesc()
	if cdd == nil {
		log.Fatal("cd is nil")
	}

	if len(cdd.ClassDetail) != 1 {
		log.Fatal("len(cdd.ClassDetail) != 1")
	}

	cd := cdd.ClassDetail[0]

	if cd.ClassName[0] != '[' {
		log.Fatal("cd.ClassName[0] != '['")
	}

	g.handleValue++

	size := int(binary.BigEndian.Uint32(g.data[0:4]))
	g.data = g.data[4:]
	//fmt.Println(fmt.Sprintf("Array size - %d", size))
	//fmt.Println("Values")

	arrayString := "["

	for i := 0; i < size-1; i++ {
		//fmt.Println(fmt.Sprintf("Index %d :", i))
		value := g.readFieldValue(cd.ClassName[1])
		arrayString += value + ", "
	}

	//fmt.Println(fmt.Sprintf("Index %d :", size-1))
	value := g.readFieldValue(cd.ClassName[1])
	arrayString += value + "]"

	return arrayString
}

func (g *GavaDeserilizer) readNewClass() {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_CLASS - 0x" + hex.EncodeToString([]byte{b1}))

	if b1 != 0x76 {
		log.Fatal("b1 != 0x76")
	}

	g.readClassDesc()

	g.handleValue++
}

func (g *GavaDeserilizer) readNewObject() *ClassDetails {
	var cdd *ClassDataDesc
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println("TC_OBJECT - 0x", hex.EncodeToString([]byte{b1}))
	if b1 != 0x73 {
		log.Fatal("Error: Illegal value for TC_OBJECT (should be 0x73)")
	}

	cdd = g.readClassDesc()

	g.handleValue++

	return g.readClassData(cdd)
}

func (g *GavaDeserilizer) readClassData(cdd *ClassDataDesc) *ClassDetails {
	//fmt.Println("classData")

	if cdd == nil {
		return nil
	}

	for classIndex := len(cdd.ClassDetail) - 1; classIndex >= 0; classIndex-- {
		cd := cdd.ClassDetail[classIndex]
		//fmt.Println(cd.ClassName)
		if g.isScSerializable(cd) {
			//fmt.Println("values")

			for _, cf := range cd.FieldDescription {
				value := g.readClassDataField(cf)
				cf.Value = value
			}
		}

		if (g.isScSerializable(cd) && g.isSCWriteMethod(cd)) || (g.isSCExternalizable(cd) && g.isSCBlockData(cd)) {
			//Start the object annotations section and indent
			//fmt.Println("objectAnnotation")
			//Loop until we have a TC_ENDBLOCKDATA
			var value = ""
			for g.data[0] != 0x78 {
				//Read a content element
				_, v := g.readContentElement()
				value += v
			}
			cd.ObjectValue = value
			//Pop and print the TC_ENDBLOCKDATA element
			g.data = g.data[1:]
			//fmt.Println("TC_ENDBLOCKDATA - 0x78")
		}
		return cd
	}
	return nil
}

func (g *GavaDeserilizer) readClassDataField(cf *ClassField) string {
	//fmt.Println(cf.Name)

	return g.readFieldValue(cf.TypeCode)
}

func (g *GavaDeserilizer) readFieldValue(typeCode byte) string {
	switch typeCode {
	case 'B': //byte
		return g.readByteField()
	case 'C': //char
		return g.readCharField()
	case 'D': //double
		return g.readDoubleField()
	case 'F': //float
		return g.readFloatField()
	case 'I': //int
		return g.readIntField()
	case 'J': //long
		return g.readLongField()
	case 'S': //short
		return g.readShortField()
	case 'Z': //boolean
		return g.readBooleanField()
	case '[': //array
		return g.readArrayField()
	case 'L': //object
		return g.readObjectField()
	default: //Unknown field type
		log.Fatal("Error: Illegal field type code ('" + string(typeCode) + "', 0x" + hex.EncodeToString([]byte{typeCode}) + ")")
	}
	return ""
}

func (g *GavaDeserilizer) readByteField() string {
	var b1 = g.data[0]
	g.data = g.data[1:]

	if int(b1) >= 0x20 && int(b1) <= 0x7e {
		//fmt.Println(fmt.Sprintf("(byte): %d", b1))
	} else {
		//fmt.Println(fmt.Sprintf("(byte): %d", b1))
	}

	return fmt.Sprintf("%d", b1)
}

func (g *GavaDeserilizer) readCharField() string {
	numBytes := g.data[0:2]
	g.data = g.data[2:]
	c1 := uint8(binary.BigEndian.Uint16(numBytes))
	//fmt.Println(fmt.Sprintf("(char): %d", c1))
	return fmt.Sprintf("%d", c1)
}

func (g *GavaDeserilizer) readDoubleField() string {
	numBytes := g.data[0:8]
	g.data = g.data[8:]
	d := float64(binary.BigEndian.Uint64(numBytes))
	//fmt.Println(fmt.Sprintf("(double): %f", d))
	return fmt.Sprintf("%f", d)
}

func (g *GavaDeserilizer) readFloatField() string {
	numBytes := g.data[0:4]
	g.data = g.data[4:]
	d := float32(binary.BigEndian.Uint32(numBytes))
	//fmt.Println(fmt.Sprintf("(float): %f", d))
	return fmt.Sprintf("%f", d)
}

func (g *GavaDeserilizer) readIntField() string {
	numBytes := g.data[0:4]
	g.data = g.data[4:]
	d := int32(binary.BigEndian.Uint32(numBytes))
	//fmt.Println(fmt.Sprintf("(int): %d", d))
	return fmt.Sprintf("%d", d)
}

func (g *GavaDeserilizer) readLongField() string {
	numBytes := g.data[0:8]
	g.data = g.data[8:]
	d := int64(binary.BigEndian.Uint64(numBytes))
	//fmt.Println(fmt.Sprintf("(long): %d", d))
	return fmt.Sprintf("%d", d)
}

func (g *GavaDeserilizer) readShortField() string {
	numBytes := g.data[0:2]
	g.data = g.data[2:]
	c1 := int8(binary.BigEndian.Uint16(numBytes))
	//fmt.Println(fmt.Sprintf("(char): %d", c1))
	return fmt.Sprintf("%d", c1)
}

func (g *GavaDeserilizer) readBooleanField() string {
	var b1 = g.data[0]
	g.data = g.data[1:]

	//fmt.Println(fmt.Sprintf("(boolean): %d", b1))
	return fmt.Sprintf("%d", b1)
}

func (g *GavaDeserilizer) readArrayField() string {
	//fmt.Println("(array)")
	switch g.data[0] {
	case 0x70:
		return g.readNullReference()
	case 0x75:
		return g.readNewArray()
	case 0x71:
		g.readPrevObject()
		return ""
	default:
		log.Fatal("Error: Unexpected array field value type")
	}
	return ""
}

func (g *GavaDeserilizer) readObjectField() string {
	//fmt.Println("(object)")
	switch g.data[0] {
	case 0x73:
		f := g.readNewObject()
		s, _ := json.Marshal(f)
		return string(s)
	case 0x71:
		g.readPrevObject()
	case 0x70:
		return g.readNullReference()
	case 0x74:
		return g.readTCString()
	case 0x76:
		g.readNewClass()
	case 0x75:
		return g.readNewArray()
	}
	return ""
}

func (g *GavaDeserilizer) isSCBlockData(cd *ClassDetails) bool {
	return cd.ClassDescFlags&0x08 == 0x08
}

func (g *GavaDeserilizer) isSCExternalizable(cd *ClassDetails) bool {
	return cd.ClassDescFlags&0x04 == 0x04
}

func (g *GavaDeserilizer) isScSerializable(details *ClassDetails) bool {
	return (details.ClassDescFlags & 0x02) == 0x02
}

func (g *GavaDeserilizer) isSCWriteMethod(cd *ClassDetails) bool {
	return cd.ClassDescFlags&0x01 == 0x01
}

func (g *GavaDeserilizer) readClassDesc() *ClassDataDesc {
	var refHandle int
	switch g.data[0] {
	case 0x72: //TC_CLASSDESC
		fallthrough
	case 0x7d: //TC_PROXYCLASSDESC
		return g.readNewClassDesc()
	case 0x70: //TC_NULL
		g.readNullReference()
		return nil
	case 0x71: //TC_REFERENCE
		refHandle = g.readPrevObject()                //Look up a referenced class data description object and return it
		for _, cdd := range g.classDataDescriptions { //Iterate over all class data descriptions
			for classIndex := 0; classIndex < len(cdd.ClassDetail); classIndex++ { //Iterate over all classes in this class data description
				if cdd.ClassDetail[classIndex].RefHandle == refHandle { //Check if the reference handle matches
					return cdd.buildClassDataDescFromIndex(classIndex) //Generate a ClassDataDesc starting from the given index and return it
				}
			}
		}
		//Invalid classDesc reference handle
		log.Fatal("Error: Invalid classDesc reference (0x" + string(refHandle))
	default:
		print("Invalid classDesc type 0x" + hex.EncodeToString([]byte{g.data[0]}))
		log.Fatal("Error illegal classDesc type.")
	}
	return nil
}
