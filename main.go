package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	goepub "github.com/bmaupin/go-epub"
	_ "github.com/imjcw/gospider/config"
	"github.com/imjcw/gospider/db"
	"github.com/imjcw/gospider/fetcher"
	"github.com/imjcw/gospider/models"
)

var storyURL string
var deamon bool
var storyPath = "./cache/"

func init() {
	flag.StringVar(&storyURL, "url", "", "链接")
	flag.BoolVar(&deamon, "deamon", false, "常驻")
}

func main() {
	flag.Parse()
	db.InitRedisHandler()

	if deamon {

	} else {
		deal(storyURL)
	}
}

func deal(storyURL string) {
	t := newTask(storyURL)
	initChapters(t)
	hash := MD5(storyURL)
	cacheDir := buildCacheDir(hash)
	if t.Image != "" {
		if resp, err := http.Get(t.Image); err == nil {
			if imageBody, err := ioutil.ReadAll(resp.Body); err == nil {
				writeFile(cacheDir+"cover.png", string(imageBody), os.O_WRONLY|os.O_CREATE)
				t.HasCover = true
			}
			resp.Body.Close()
		}
	}
	for _, chapter := range t.Chapters {
		times := 0
		if db.RedisHandler.SIsMember(hash, chapter.Hash).Val() {
			fmt.Printf("《%s》, 重试次数: %d, %s\n", t.Title, times, chapter.Title)
			continue
		}
		content := ""
		for {
			fmt.Printf("《%s》, 重试次数: %d, %s", t.Title, times, chapter.Title)
			times++
			doc, err := fetcher.Fetch(chapter.URL)
			if err != nil {
				time.Sleep(2 * time.Second)
				fmt.Println(", Error: 请求失败: " + err.Error())
				continue
			}
			if doc == nil {
				fmt.Println(", Error: 解析页面DOM失败")
				continue
			}
			content = ""
			selection := doc.Find(t.Rule.ChapterDetailDom)
			if selection == nil {
				fmt.Println(", Error: 解析内容DOM失败")
				continue
			}
			fmt.Println("")
			selection.Each(func(i int, s *goquery.Selection) {
				line, err := s.Html()
				if err != nil {
					return
				}
				content += "<p>" + strings.TrimLeft(line, "　") + "</p>\n"
			})
			if content != "" || times > t.Rule.RetryTimes {
				content = "<h2>" + chapter.Title + "</h2>\n" + content
				break
			}
			time.Sleep(1 * time.Second)
		}
		db.RedisHandler.SAdd(hash, chapter.Hash)
		writeFile(fmt.Sprintf("%s%s.html", cacheDir, chapter.Hash), content, os.O_WRONLY|os.O_CREATE)
	}

	fmt.Println("合并文件中...")
	epub := goepub.NewEpub(t.Title)
	epub.SetAuthor(t.Author)
	epub.SetDescription(t.Description)
	if t.HasCover {
		coverImagePath, _ := epub.AddImage(cacheDir+"cover.png", "cover.png")
		epub.SetCover(coverImagePath, "")
	}

	for _, chapter := range t.Chapters {
		// 读取文件内容
		content, err := os.ReadFile(fmt.Sprintf("%s%s.html", cacheDir, chapter.Hash))
		if err != nil {
			panic(err)
		}
		epub.AddSection(string(content), chapter.Title, "", "")
	}
	if err := epub.Write(storyPath + "epubs/" + t.Title + ".epub"); err != nil {
		panic(err)
	}
	fmt.Println("文件生成成功, " + storyPath + "epubs/" + t.Title + ".epub")
}

func newTask(storyURL string) *models.Task {
	if storyURL == "" {
		panic("url is needed.")
	}
	pURL, err := url.Parse(storyURL)
	if err != nil {
		panic(err)
	}
	baseURL := ""
	if pURL.Scheme != "" {
		baseURL += pURL.Scheme + "://"
	}
	if pURL.Host != "" {
		baseURL += pURL.Host
	}
	if pURL.Port() != "" {
		baseURL += ":" + pURL.Port()
	}
	return &models.Task{
		URL:     storyURL,
		BaseURL: baseURL,
		Rule: models.Rule{
			RetryTimes:           30,
			StoryTitleDom:        "#maininfo #info h1",
			StoryAuthorDom:       "#maininfo #info p",
			StoryImageDom:        "#sidebar #fmimg img",
			StoryDescriptionDom:  "#maininfo #intro p",
			ChapterListDom:       "#list dt, #list dd",
			ChapterDetailDom:     "#content p",
			ChapterListStartFrom: "正文",
		},
	}
}

func buildCacheDir(hash string) (cacheDir string) {
	cacheDir = storyPath + hash + "/"
	_, _err := os.Stat(cacheDir)
	if _err == nil {
		return
	}
	if !os.IsNotExist(_err) {
		panic(_err)
	}
	if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
		panic(err)
	}
	return
}

func writeFile(path, content string, mode int) {
	file, err := os.OpenFile(path, mode, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	write := bufio.NewWriter(file)
	//及时关闭file句柄
	defer file.Close()
	write.WriteString(content)
	write.Flush()
}

func initChapters(t *models.Task) {
	doc, _ := fetcher.Fetch(t.URL)
	startLog := false
	storyTitleDom := doc.Find(t.Rule.StoryTitleDom)
	if storyTitleDom == nil {
		panic("story title dom undefined")
	}
	storyAuthorDom := doc.Find(t.Rule.StoryAuthorDom)
	if storyAuthorDom == nil {
		panic("story author dom undefined")
	}
	storyImageDom := doc.Find(t.Rule.StoryImageDom)
	if storyImageDom == nil {
		panic("story image dom undefined")
	}
	storyDescriptionDom := doc.Find(t.Rule.StoryDescriptionDom)
	if storyDescriptionDom == nil {
		panic("story description dom undefined")
	}
	t.Title = doc.Find(t.Rule.StoryTitleDom).Text()
	t.Author = doc.Find(t.Rule.StoryAuthorDom).First().Text()
	t.Author = strings.Split(t.Author, "：")[1]
	t.Image = doc.Find(t.Rule.StoryImageDom).AttrOr("src", "")
	t.Description = doc.Find(t.Rule.StoryDescriptionDom).Text()
	doc.Find(t.Rule.ChapterListDom).Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		if t.Rule.ChapterListStartFrom != "" && strings.LastIndex(html, t.Rule.ChapterListStartFrom) == -1 && !startLog {
			return
		}
		startLog = true
		aDom := s.Find("a")
		link, exist := aDom.Attr("href")
		if !exist {
			return
		}
		if strings.Index(link, "h") != 0 {
			link = t.BaseURL + "/" + strings.TrimLeft(link, "/")
		}
		t.Chapters = append(t.Chapters, &models.Chapter{
			Title: aDom.Text(),
			URL:   link,
			Hash:  MD5(link),
		})
	})
	return
}

// MD5 .
func MD5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
