package cloudstore

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Endpoint reflects a common interface for structs with connection information
// to an arbitrary |FileSystem|.
type Endpoint interface {
	// CheckPermissions connects to the endpoint and confirms that the
	// passed credentials have read/write permissions in the root directory.
	CheckPermissions() error
	// Connect returns a FileSystem to be used by the caller, allowing the caller
	// to specify an arbitrary set of additional |Properties|. In most cases,
	// this will be unnecessary, as all connection details will be specified
	// by the Endpoint. |Properties| passed will be merged with those defined in the
	// Endpoint, overwriting the Endpoint properties where necessary.
	Connect(Properties) (FileSystem, error)
	// Validate inspects the endpoint and confirms that all internal fields are
	// well-formed. Also satisfies the Model interface.
	Validate() error
}

// UnmarshalEndpoint takes a byte array of json data (usually from etcd) and
// returns the appropriate |Endpoint| interface implementation.
func UnmarshalEndpoint(data []byte) (ep Endpoint, err error) {
	var base BaseEndpoint
	if err = json.Unmarshal(data, &base); err != nil {
		return
	}
	switch base.Type {
	case "s3":
		ep = new(S3Endpoint)
	case "sftp":
		ep = new(SFTPEndpoint)
	case "gcs":
		ep = new(GCSEndpoint)
	default:
		panic(fmt.Sprintf("unknown endpoint type: %s", base.Type))
	}
	err = json.Unmarshal(data, &ep)
	return
}

// BaseEndpoint provides common fields for all endpoints. Though it currently
// only contains a |Name| field, it's important to maintain this inheritence
// to allow us to use |Name| as a primary key in the endpoint namespace.
type BaseEndpoint struct {
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	PermissionTestFilename string `json:"permission_test_filename"`
}

// Validate satisfies the Model interface from model-builder. Endpoint implementations
// are built from SQL, and Validate()'d as they're ETL'd into etcd.
func (ep *BaseEndpoint) Validate() error {
	if ep.Name == "" {
		return errors.New("must specify an endpoint name")
	}
	return nil
}
