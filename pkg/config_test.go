package pkg

// "bytes"
// "encoding/json"
// "io/fs"
// "os"
// "path/filepath"
// "testing"

// "github.com/stretchr/testify/assert"
// "github.com/stretchr/testify/require"

// func TestConfig(t *testing.T) {
// 	assert := assert.New(t)

// 	co := User{"oluwamuyiwa", "onigbs@gmail.com"}

// 	err := Config(co)

// 	if assert.NoError(err, "Could not evn run config successfully: %v", err) {
// 		if dir, err := os.UserCacheDir(); assert.Nil(err, "Error opening UserCache: %v", err) {
// 			b, err := fs.ReadFile(os.DirFS(filepath.Join(dir, ".git")), ".config")
// 			if assert.NoError(err, "Could not read file system: %v", err) {
// 				buf := bytes.NewReader(b)
// 				dec := json.NewDecoder(buf)
// 				var conf User
// 				if err := dec.Decode(&conf); assert.NoError(err, "error while decoding json") {
// 					require.Equal(t, co, conf, "What is stored is not what we read")
// 				}
// 			}
// 		}
// 	}
// }
