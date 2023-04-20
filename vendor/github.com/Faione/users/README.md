# Users

Package users retrieves Users in the Linux system

## Install

```shell
go get github.com/Faione/users
```

## Usage

```go
loginUsers, err := ListLogged()

if err != nil {
    return
}

rlt, _ := json.MarshalIndent(loginUsers, "", "     ")
fmt.Println(string(rlt))
```