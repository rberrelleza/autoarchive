package rest

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
)

func GetDescriptor(url string) (*model.Descriptor, error) {
	var d model.Descriptor

	res, err := http.Get(url)
	if err != nil {
		return &d, err
	}

	err = checkJSONResponse(res, 200)
	if err != nil {
		return &d, err
	}

	return model.DecodeDescriptor(res.Body)
}
