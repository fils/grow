package operations

import (
	"context"
	"fmt"
	"net/http"

	minio "github.com/minio/minio-go/v7"
)

// ObjectExists simply attepts to stat an object to allow us to know if it exists or not
// we could return the stat object, but don't have a need for it now
func ObjectExists(mc *minio.Client, w http.ResponseWriter, r *http.Request, bucket, object string) error {

	_, err := mc.StatObject(context.Background(), bucket, object, minio.StatObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
