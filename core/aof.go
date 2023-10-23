package core

import (
	"fmt"
	"go-database/config"
	"log"
	"os"
	"strings"
)

func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Print("error", err)
		return
	}
	log.Println("rewriting AOF file at", config.AOFFile)
	for k, obj := range dataStore {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF file rewrite complete")
}

func dumpKey(fp *os.File, k string, obj *Obj) {
	//get the command
	cmd := fmt.Sprintf("SET %s %s", k, obj.value)
	tokens := strings.Split(cmd, "")
	fp.Write(Encode(tokens, false))
}
