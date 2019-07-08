package helpers

type redisError interface {
	Err() error
}

func IsRedisError(err redisError) bool {
	//

	return err.Err() != nil
}
