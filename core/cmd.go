package core

type RedisCmd struct {
	Cmd  string
	Args []string
}

type RedisCmds []*RedisCmd // this will be used to accept multiple commands <command pipelining>
