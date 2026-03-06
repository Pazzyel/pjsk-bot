package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"pjsk-bot/server/config"
	"regexp"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type PJSKService struct {
	pjskConfig *config.PJSKConfig
}

func NewPJSKService(cfg *config.PJSKConfig) *PJSKService {
	p := &PJSKService{cfg}
	p.pjskConfig = cfg
	return p
}

func (p *PJSKService) GetCharts(id string, level string) ([]byte, error) {
	// 先检查文件是否存在
	dir := p.pjskConfig.PJSK.Charts.SavePath
	// 创建目录（如果不存在）
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, errors.New("创建目录失败: " + err.Error())
	}
	// 文件名格式：{id}{level}.png，例如 001exp.png
	fileName := id + level + ".png"
	path := filepath.Join(dir, fileName)
	exists, err := checkFile(path)
	if err != nil {
		return nil, err
	}
	if exists {
		// 读取文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// 文件不存在则从网络获取
	if !checkLevel(level) {
		return nil, errors.New("难度错误，仅支持easy,normal,hard,expert,append，必须拼写正确")
	}

	// 处理谱面图
	data, err := http.Get(p.pjskConfig.PJSK.Charts.RequestPath + id + "/" + level + ".svg")
	if err != nil {
		return nil, err //ors.New("获取谱面图失败")
	}
	defer data.Body.Close()

	// 处理404
	if data.StatusCode == 404 {
		return nil, errors.New("未找到该谱面，请检查ID和难度是否正确")
	}

	svg, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, err //ors.New("读取谱面图失败")
	}

	//转换SVG，替换原版的紫色背景
	svgStr := string(svg)
	svgStr = replaceCSSColor(svgStr, ".lane", "#555555")
	svgStr = replaceCSSColor(svgStr, ".background", "#333333")
	svgStr = replaceCSSColor(svgStr, ".meta", "#000000")

	//把SVG渲染成PNG字节流
	icon, err := oksvg.ReadIconStream(bytes.NewReader([]byte(svgStr)))
	if err != nil {
		return nil, errors.New("渲染PNG图片失败")
	}

	width := int(icon.ViewBox.W)
	height := int(icon.ViewBox.H)
	image := image.NewRGBA(image.Rect(0, 0, width, height))
	scanner := rasterx.NewScannerGV(width, height, image, image.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)
	icon.Draw(raster, 1.0)

	// 编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		return nil, err //ors.New("合并图片失败")
	}
	imageBytes := buf.Bytes()

	// 将图片保存到本地
	err = os.WriteFile(path, imageBytes, 0644)
	if err != nil {
		return nil, errors.New("保存图片失败")
	}

	return imageBytes, nil
}

// 检查难度
func checkLevel(level string) bool {
	switch level {
	case "easy":
		return true
	case "normal":
		return true
	case "hard":
		return true
	case "expert":
		return true
	case "mater":
		return true
	case "append":
		return true
	default:
		return false
	}
}

// 检查文件是否存在
func checkFile(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		fmt.Println("文件存在:", path)
		return true, nil
	} else if os.IsNotExist(err) {
		fmt.Println("文件不存在:", path)
		return false, nil
	} else {
		fmt.Println("检查文件出错:", err)
		return false, err
	}
}

// 把svg文件的指定标签的fill值替换成指定
func replaceCSSColor(svg, selector, color string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(%s\s*\{[^}]*?)fill\s*:\s*#[0-9a-fA-F]{3,8}([^}]*\})`, regexp.QuoteMeta(selector)))
	if re.MatchString(svg) {
		return re.ReplaceAllString(svg, fmt.Sprintf("${1}fill: %s${2}", color))
	}
	return svg
}

// 把源图片透明像素转为黑色
func transparentToBlack(srcPtr *image.RGBA) {
	src := *srcPtr
	bounds := src.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			if c.A == 0 {
				// 透明像素 → 黑色
				src.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
}

// 获取歌曲封面
func (p *PJSKService) GetJackets(id string) ([]byte, error) {
	// 先检查文件是否存在
	dir := p.pjskConfig.PJSK.Jackets.SavePath
	// 创建目录（如果不存在）
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, errors.New("创建目录失败: " + err.Error())
	}

	// 文件名格式：{id}.png，例如 001.png
	fileName := id + ".webp"
	path := filepath.Join(dir, fileName)
	exists, err := checkFile(path)
	if err != nil {
		return nil, err
	}

	if exists {
		// 读取文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// 文件不存在则从网络获取
	idFormat := "jacket_s_" + id
	data, err := http.Get(p.pjskConfig.PJSK.Jackets.RequestPath + idFormat + "/" + idFormat + ".webp")
	if err != nil {
		return nil, err //ors.New("获取封面图失败")
	}
	defer data.Body.Close()

	// 处理404
	if data.StatusCode == 404 {
		return nil, errors.New("未找到该封面，请检查ID是否正确")
	}
	jacket, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, err //ors.New("读取封面图失败")
	}

	// 将图片保存到本地
	err = os.WriteFile(path, jacket, 0644)
	if err != nil {
		return nil, errors.New("保存图片失败")
	}
	return jacket, nil
}
