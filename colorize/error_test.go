package colorize

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func colorerToFormat(name string) func(format string, a ...interface{}) string {
	return func(format string, a ...interface{}) string {
		return fmt.Sprintf("<"+name+">"+format+"</"+name+">", a...)
	}
}

func init() {
	redColorFun = colorerToFormat("red")
	cyanColorFun = colorerToFormat("cyan")
	greenColorFun = colorerToFormat("green")
}

func Test_NewError(t *testing.T) {
	cErr := NewError("example %s of %s color %s error %s!", Cyan("value1"), Red("value2"), Green("value3"), None("value4"))
	require.Equal(t, "example 'value1' of value2 color value3 error value4!", cErr.Error())
	require.Equal(t, "example <cyan>value1</cyan> of <red>value2</red> color <green>value3</green> error value4!", cErr.ColorError())
}

func Test_PathComponent(t *testing.T) {
	err := errors.New("normal error")
	cErr2 := NewEntityError("strategy %s", "strategy-name").SetSubError(err)
	cErr := NewPathError("some.path[0]", cErr2)
	require.Equal(t, "path 'some.path[0]': strategy 'strategy-name': normal error", cErr.Error())
	require.Equal(t, "path <cyan>some.path[0]</cyan>: strategy <cyan>strategy-name</cyan>: normal error", cErr.ColorError())

	require.True(t, HasPathComponent(cErr))
	require.False(t, HasPathComponent(cErr2))
	require.False(t, HasPathComponent(err))

	err = RemovePathComponent(err)
	err1 := RemovePathComponent(cErr)
	err2 := RemovePathComponent(cErr2)
	require.False(t, HasPathComponent(err))
	require.False(t, HasPathComponent(err1))
	require.False(t, HasPathComponent(err2))
}

func Test_ResponseDb_Error(t *testing.T) {
	tail := []Part{
		None("\n\n   diff (--- expected vs +++ actual):\n"),
	}
	tail = append(tail, MakeColorDiff([]string{"1", "2", "3"}, []string{"3", "2", "1"})...)

	cErr2 := NewEntityNotEqualError("quantity of %s does not match:", "items in database", 12, 13).AddParts(tail...)
	cErr := NewPathError("some.test[path]", cErr2)

	require.Equal(t, `path 'some.test[path]': quantity of 'items in database' does not match:
     expected: 12
       actual: 13

   diff (--- expected vs +++ actual):
+3
+2
 1
-2
-3
`, cErr.Error())
	require.Equal(t, `path <cyan>some.test[path]</cyan>: quantity of <cyan>items in database</cyan> does not match:
     expected: <green>12</green>
       actual: <red>13</red>

   diff (--- expected vs +++ actual):
<red>+3
+2
</red> 1
<green>-2
-3
</green>`, cErr.ColorError())
}

func Test_Mocks_Error(t *testing.T) {
	dump := None("%dump%")
	cErr2 := NewEntityNotEqualError("different value %s", "var-name", 34, 56)
	cErr := NewEntityError("request constraint %s", "some-name").SetSubError(cErr2).AddPostfix(
		None(", request was:\n\n"), dump,
	)
	require.Equal(t, "request constraint 'some-name': different value 'var-name'\n     expected: 34\n       actual: 56, request was:\n\n%dump%", cErr.Error())
	require.Equal(t, "request constraint <cyan>some-name</cyan>: different value <cyan>var-name</cyan>\n     expected: <green>34</green>\n       actual: <red>56</red>, request was:\n\n%dump%", cErr.ColorError())
}
