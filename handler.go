package main

import (
	"sync"
)

var Handlers = map[string]func([]Value) Value{
	"COMMAND": command,
	"PING": ping,
	"SET":  Set,
	"GET":  Get,
	"HGET": HGet,
	"HSET": HSet,
	"HGETALL": HGetAll,
}

func command(value []Value) Value {
	return EmptyValue()
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func Set(args []Value) Value {
	if len(args) != 2 {
		return ErrValue("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return OkValue()
}

func Get(args []Value) Value {
	if len(args) != 1 {
		return ErrValue("ERR wrong number of arguments for 'get' command")
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return NullValue()
	}

	return BulkValue(value)
}

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func HSet(args []Value) Value {
	if len(args) != 3 {
		return ErrValue("ERR wrong number of arguments for 'hset' command")
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMu.Unlock()

	return OkValue()
}

func HGet (args []Value) Value{
	if len(args) != 2 {
		return ErrValue("ERR wrong number of arguments for 'hget' command")
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMu.RLock()
	value, ok := HSETs[hash][key]
	HSETsMu.RUnlock()

	if !ok {
		return NullValue()
	}

	return BulkValue(value)
}

func HGetAll (args []Value) Value {
	if len(args) != 1 {
		return ErrValue("ERR wrong number of arguments for 'hgetall' command")
	}

	hash := args[0].bulk

	HSETsMu.RLock()
	values, ok := HSETs[hash]
	HSETsMu.RUnlock()

	if !ok {
		return NullValue()
	}

	valuesResult := make([]Value, 0, len(values))

	for _,value := range values {
		valuesResult = append(valuesResult, BulkValue(value))
	}

	return ArrayValue(valuesResult)
}
