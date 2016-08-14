package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"	
	"strings"
	"io"
	"bufio"
)
var (
	SPLIT 	= "|"
	W_CMD_I = "w_i"   		//协商写服务器 c->s
	W_CMD_N = []byte("w_n") //请求下一行 s->c
	W_CMD_C = "w_c"			//写服务器结束 c->s
	R_CMD_I = "r_i"			//协商读服务器 c->s
	R_CMD_N = "r_n"			//请求读下一行 c->s
	R_CMD_C = []byte("r_c")	//读服务器结束 s->c
	ERR_CMD = []byte("err_cmd")
	fileName string
)

func init() {
	fileName = getCurrentDirectory()+"/aoki_sp.data"
}

// HandleError 错误处理
func HandleError(err error) {
	fmt.Println("occurred error:", err)
    panic("error.........")
}

//获取当前路径
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		HandleError(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func writeFile(data []string) {
	file,err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		HandleError(err)
	}
	for _, line := range data {
		file.WriteString(line)
	}
}

func readFile() []string {
	data := make([]string, 0)
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		HandleError(err)
	}
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			HandleError(err)
		}		
		data = append(data,line)		
	}	
	return data
}

func writeCmd(conn net.Conn) {
	data := make([]string,1)
	buf := make([]byte, 1024)
	for {
		c, err := conn.Write(W_CMD_N)
		if err != nil {
			HandleError(err)
		}
		// println("write server, send W_CMD_N")
		c, err = conn.Read(buf)
		if err != nil {
			HandleError(err)
		}
		line := string(buf[:c])		
		// println("line=",line)
		if line == W_CMD_C {			
			break
		}
		data = append(data, line)		
	}
	writeFile(data)
	println("write success.")
}

func readCmd(conn net.Conn) {
	data := readFile()
	buf := make([]byte, 16)
	for _, line := range data {
		if line == "" {
			println("line is nil")
			continue
		}
		c, err := conn.Write([]byte(line))
		if err != nil {
			HandleError(err)
		}
		c, err = conn.Read(buf)
		if err != nil {
			HandleError(err)
		}
		cmd := string(buf[:c])
		if cmd != R_CMD_N {
			println("read server,cmd error! cmd=", cmd)
			return
		}
	}
	println("read success.")
	_, err := conn.Write(R_CMD_C)
	if err != nil {
		HandleError(err)
	}
}

// func write(conn net.Conn, data string) {
// 	conn.Write([]byte(data+SPLIT))
// }

// func read(conn net.Conn) string {
// 	data := make([]byte, 1024)

// }

func handleConn(conn net.Conn) {
	println("run handleConn .......")
	defer conn.Close()
	buf := make([]byte, 8)
	c, err := conn.Read(buf)
	if err != nil {
		HandleError(err)
	}

	cmd := string(buf[:c])
	if cmd == W_CMD_I {
		writeCmd(conn)
	} else if cmd == R_CMD_I {
		readCmd(conn)
	} else {
		conn.Write(ERR_CMD)
	}
}

//initial listener and run
func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6213")
	if err != nil {
		HandleError(err)
	}
	defer listener.Close()
	println("running ...")	

	for {
		conn, err := listener.Accept()		
		if err != nil {
			HandleError(err)
		}
		// println("accept conn: ",conn)
		go handleConn(conn)
	}	
}