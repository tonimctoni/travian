package main

import "regexp"
import "errors"
import "strconv"
import "bytes"

type TravianData struct{
    capacity int64
    wood int64
    clay int64
    iron int64
    korn_capacity int64
    korn int64
    free_korn int64

    is_logged_in bool
    is_upgrading bool
}

var find_dorf1_capacity *regexp.Regexp
var find_dorf1_wood *regexp.Regexp
var find_dorf1_clay *regexp.Regexp
var find_dorf1_iron *regexp.Regexp
var find_dorf1_korn_capacity *regexp.Regexp
var find_dorf1_korn *regexp.Regexp
var find_dorf1_free_korn *regexp.Regexp
func (t *TravianData) gather_data(content []byte) error{
    t.is_logged_in=!bytes.Contains(content, []byte("<h2>Willkommen auf der Welt"))
    if !t.is_logged_in{
        return errors.New("Not logged in")
    }
    t.is_upgrading=bytes.Contains(content, []byte("<h5>Bauauftr"))

    if find_dorf1_capacity==nil{
        find_dorf1_capacity=regexp.MustCompile("<span class=\"value\" id=\"stockBarWarehouse\">([0-9]*)</span>")
    }
    if find_dorf1_wood==nil{
        find_dorf1_wood=regexp.MustCompile("<span id=\"l1\" class=\"value\">([0-9]*)</span>")
    }
    if find_dorf1_clay==nil{
        find_dorf1_clay=regexp.MustCompile("<span id=\"l2\" class=\"value\">([0-9]*)</span>")
    }
    if find_dorf1_iron==nil{
        find_dorf1_iron=regexp.MustCompile("<span id=\"l3\" class=\"value\">([0-9]*)</span>")
    }
    if find_dorf1_korn_capacity==nil{
        find_dorf1_korn_capacity=regexp.MustCompile("<span class=\"value\" id=\"stockBarGranary\">([0-9]*)</span>")
    }
    if find_dorf1_korn==nil{
        find_dorf1_korn=regexp.MustCompile("<span id=\"l4\" class=\"value\">([0-9]*)</span>")
    }
    if find_dorf1_free_korn==nil{
        find_dorf1_free_korn=regexp.MustCompile("<span id=\"stockBarFreeCrop\" class=\"value\">([0-9]*)</span>")
    }

    use_regex:=func(re *regexp.Regexp) (int64, error){
        matches:=re.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return 0, errors.New("len(matches)!=1")
        }

        if len(matches[0])!=2{
            return 0, errors.New("len(matches[0])!=2")
        }

        if len(matches[0][1])<=0{
            return 0, errors.New("len(matches[0][1])<=0")
        }

        ret, err:=strconv.ParseInt(string(matches[0][1]), 10, 64)
        if err!=nil{
            return 0, err
        }

        return ret, nil
    }

    var err error
    t.capacity, err=use_regex(find_dorf1_capacity)
    if err!=nil{
        return err
    }
    t.wood, err=use_regex(find_dorf1_wood)
    if err!=nil{
        return err
    }
    t.clay, err=use_regex(find_dorf1_clay)
    if err!=nil{
        return err
    }
    t.iron, err=use_regex(find_dorf1_iron)
    if err!=nil{
        return err
    }
    t.korn_capacity, err=use_regex(find_dorf1_korn_capacity)
    if err!=nil{
        return err
    }
    t.korn, err=use_regex(find_dorf1_korn)
    if err!=nil{
        return err
    }
    t.free_korn, err=use_regex(find_dorf1_free_korn)
    if err!=nil{
        return err
    }

    return nil
}