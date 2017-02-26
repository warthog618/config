package json

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "strings"
)

// encoding/json
//bool, for JSON booleans
//float64, for JSON numbers
//string, for JSON strings
//[]interface{}, for JSON arrays
//map[string]interface{}, for JSON objects
//nil for JSON null

// More general reader types
// cfgconv package will perform these conversions.
//bool,int,uint, for booleans
//float64, for float
//float64,int, for int
//float64,uint for uint
//string, for strings
//[]interface{}, for arrays
//map[string]interface{}, for objects
//nil for JSON null

type reader struct {
  config map[string]interface{}
}

func (r * reader)Contains(key string) (bool) {
  _,ok := r.Read(key)
  return ok
}

func (r * reader)Read(key string) (interface{},bool) {
  if val,err := getFromMapTree(r.config,key,"."); err == nil {
    return val,true
  } else {
    return struct{}{},false
  }
}

func NewBytes(cfg []byte) (*reader,error){
  var config map[string]interface{}
	err := json.Unmarshal(cfg, &config)
	if err != nil {
    return &reader{},err
	}
  return &reader{config},nil
}

func NewFile(filename string) (*reader,error){
  cfg,err := ioutil.ReadFile(filename)
  if err != nil {
    return &reader{},err
  }
  return NewBytes(cfg)
}

func getFromMapTree(node map[string]interface{}, key string, pathSep string) (interface{},error) {
  // full key match
  if value,ok := node[key]; ok {
    return value,nil
  }
  // nested path match
  path := strings.Split(key,pathSep)
  if value,ok := node[path[0]]; ok {
    if len(path) == 1 {
      return value,nil
    } else {
      switch v := value.(type) {
      case map[string]interface{}:
        return getFromMapTree(v,strings.Join(path[1:],pathSep),pathSep)
      }
    }
  }
  // no match
  return nil,fmt.Errorf("key '%v' not found", key)
}
