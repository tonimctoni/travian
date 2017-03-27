package main

import "net/http"
import "net/url"
import "io/ioutil"
import "regexp"
import "errors"
import "net/http/cookiejar"
import "math/rand"
import "time"

type TravianClient struct{
    http.Client
}

func NewTravianClient() TravianClient {
    cookie_jar, err:=cookiejar.New(nil)
    if err!=nil{
        panic(err.Error())
    }
    return TravianClient{http.Client{Jar: cookie_jar}}
}

func (t *TravianClient) Get_and_wait(url string) (resp *http.Response, err error){
    resp,err=t.Get(url)
    time.Sleep(time.Duration(rand.Int63n(1500)+500)*time.Millisecond)
    return
}

func (t *TravianClient) PostForm_and_wait(url string, data url.Values) (resp *http.Response, err error){
    resp,err=t.PostForm(url, data)
    time.Sleep(time.Duration(rand.Int63n(1500)+500)*time.Millisecond)
    return
}

var find_login *regexp.Regexp
func (t *TravianClient) login(name, password string) []byte{
    //Get main page content
    content:=func() []byte{
        response, err:=t.Get_and_wait("http://ts4.travian.de")
        if err!=nil{
            panic(err.Error())
        }
        content, err:=ioutil.ReadAll(response.Body)
        if err!=nil {
            panic(err.Error())
        }
        return content
    }()

    //Get login var from main page content
    login_var, err:=func(content []byte) (string,error) {
        if find_login==nil{
            // fmt.Println("Initializing find_login re")
            find_login=regexp.MustCompile("<input type=\"hidden\" name=\"login\" value=\"([0-9]*)\" />")
        }

        matches:=find_login.FindAllSubmatch(content,-1)
        if len(matches)!=1{
            return "", errors.New("len(matches)!=1")
        }

        if len(matches[0])!=2{
            return "", errors.New("len(matches[0])!=2")
        }

        if len(matches[0][1])<=0{
            return "", errors.New("len(matches[0][1])<=0")
        }

        return string(matches[0][1]), nil
    }(content)
    if err!=nil{
        panic(err.Error())
    }

    return func() []byte{
        response, err:=t.PostForm_and_wait("http://ts4.travian.de/dorf1.php", url.Values{
            "name": {name},
            "password": {password},
            "s1": {"Einloggen"},
            "w": {"1366:768"},
            "login": {login_var},
            "lowRes": {"0"},
            })

        dorf1_content, err:=ioutil.ReadAll(response.Body)
        if err!=nil {
            panic(err.Error())
        }
        return dorf1_content
    }()
}