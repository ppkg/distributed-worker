package util

import (
	"fmt"

	"google.golang.org/grpc/status"
)

func ConvertGrpcError(err error) error {
	grpcErr, ok := status.FromError(err)
	if !ok {
		return err
	}
	return fmt.Errorf("code:%d,detail:%s", grpcErr.Code(), grpcErr.Proto().Message)
}
