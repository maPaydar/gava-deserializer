package test

import (
	"bufio"
	"os"
	"testing"

	"github.com/maPaydar/gava-deserializer"
	"github.com/stretchr/testify/assert"
)

func readLine(path string) string {
	inFile, err := os.Open(path)
	if err != nil {
		//fmt.Println(err.Error() + `: ` + path)
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
