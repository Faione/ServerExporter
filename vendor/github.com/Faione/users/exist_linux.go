package users

import (
	"bufio"
	"errors"
	"os"
	"os/user"
	"strings"
)

type UserFilter func(user *user.User) bool

// 读取 `/etc/passwd` 文件，获得所有用户的信息
//
// 可以通过设置 UserFilter 来对用户进行过滤
func ListAll(filters ...UserFilter) ([]*user.User, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var users []*user.User
	bf := bufio.NewReader(file)
LOOP:
	for {
		line, _, err := bf.ReadLine()
		if err != nil {
			break
		}

		user, err := parsePasswd(string(line))
		if err != nil {
			continue
		}

		for _, filter := range filters {
			if !filter(user) {
				continue LOOP
			}
		}

		users = append(users, user)
	}

	return users, nil

}

func parsePasswd(passwd string) (*user.User, error) {
	cols := strings.Split(passwd, ":")
	if len(cols) != 7 {
		return nil, errors.New("unexpected number of fields in /etc/passwd")
	}

	return &user.User{
		Username: cols[0],
		Uid:      cols[2],
		Gid:      cols[3],
		Name:     cols[4],
		HomeDir:  cols[5],
	}, nil
}
