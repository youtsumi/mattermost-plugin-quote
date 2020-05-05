package plugin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)
	p := SharePostPlugin{}

	api := &plugintest.API{}
	api.On("LogDebug", GetMockArgumentsWithType("string", 7)...).Return()
	defer api.AssertExpectations(t)
	p.SetAPI(api)

	p.router = p.InitAPI()
	p.setConfiguration(&configuration{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	p.ServeHTTP(nil, w, r)

	result := w.Result()
	defer result.Body.Close()

	assert.NotNil(result)
	bodyBytes, err := ioutil.ReadAll(result.Body)
	assert.Nil(err)
	bodyString := string(bodyBytes)

	assert.Equal("Installed SharePostPlugin v0.1.0", bodyString)
}

func GetMockArgumentsWithType(typeString string, num int) []interface{} {
	ret := make([]interface{}, num)
	for i := 0; i < len(ret); i++ {
		ret[i] = mock.AnythingOfTypeArgument(typeString)
	}
	return ret
}
