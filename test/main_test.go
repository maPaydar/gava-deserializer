package test

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/maPaydar/gava-deserializer"
	"github.com/maPaydar/gava-deserializer/pkg"
	"github.com/stretchr/testify/assert"
)

func readLine(path string) string {
	inFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error() + `: ` + path)
		return ""
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func Test(t *testing.T) {
	inBytes := []byte(readLine("./test.txt"))

	g := gava.NewGavaDeserilizer(inBytes)
	parsedObject := g.Parse()

	assert.NotNil(t, parsedObject)
	assert.Equal(t, len(parsedObject.FieldDescription), 3)
}

func TestHex(t *testing.T) {
	hexB := "aced00057372002f696d2e6163746f722e7365727665722e6469616c6f672e47726f75704469616c6f675374617465536e617073686f74000000000000000002000449000767726f757049644c000f6c6173744d657373616765446174657400134c6a6176612f74696d652f496e7374616e743b4c000c6c617374526561644461746571007e00014c000f6c617374526563656976654461746571007e00017870000000007372000d6a6176612e74696d652e536572955d84ba1b2248b20c00007870770d02000000005b1bd9bb352ad700787371007e0003770d02000000005b056ab729f63000787371007e0003770d02000000005b1bd9bb352ad70078"
	data := pkg.DecodeHex(hexB)

	g := gava.NewGavaDeserilizer(data)
	parsedObject := g.Parse()

	assert.NotNil(t, parsedObject)
	assert.Equal(t, len(parsedObject.FieldDescription), 4)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
