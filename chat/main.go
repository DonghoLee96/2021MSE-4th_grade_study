package main

import (
	"flag"          // 명령줄 옵션을 편리하게 사용하기 위한 패키지
	"log"           // 로그 패키지
	"net/http"      // 웹 서버, 클라이언트 관련 패키지
	"os"            // 표준 스트림을 활용하기 위한 패키지
	"path/filepath" // 외부파일을 가져오기 위한 패키지
	"sync"          // sync.Once 를 사용하기 위한 패키지
	"text/template" // 템플릿 컴파일 위한 패키지

	"github.com/matryer/goblueprints/chapter1/trace" // trace 관련 패키지
)

// 템플릿을 로드, 컴파일, 전달하는 구조체
type templateHandler struct {
	once     sync.Once          // 특정 함수를 1 번만 실행시키고자 할 때 사용하는 타입, (여러 go 루틴에서 실행한다고 해도 해당 함수는 1번만 실행됨)
	filename string             // .html 파일 이름
	templ    *template.Template // templ 은 하나의 템플릿을 의미
}

// ServeHTTP 는 HTTP 요청을 처리함 ( 소스파일 로드 -> 템플릿 컴파일 후 실행 -> 지정된 http.ResponseWriter 메소드 출력 작성)
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// templates 폴더 안에 t.filename html 템플릿을 한번만 컴파일 해주고 실행
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":9080", "The addr of the application.") // 문자열 플래그 선언 (이름, 값, 용도)
	flag.Parse()                                                            // 플래그가 모두 선언되면 커맨드 라인 파싱

	r := newRoom()                  // 채팅방을 만듦
	r.tracer = trace.New(os.Stdout) // 결과를 os.Stdout 표준 출력 파이프로 보낼 객체 생성 (room.go에서 활용)

	// chat.html 템플릿을 연 다음 요청에 대해 어떤 핸들러를 사용할지 지정하는 라우팅 역할
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// 백그라운드 프로세스에서 채팅방이 동작하기 시작
	go r.run()

	// ListenAndServe 는 지정된 포트에 웹 서버를 열어줌.
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
