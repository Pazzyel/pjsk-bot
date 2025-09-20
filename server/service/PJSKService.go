package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"server/config"
)

type PJSKService struct {
	pjskConfig *config.PJSKConfig
}

func (p *PJSKService) Construct(cfg *config.PJSKConfig) {
	p.pjskConfig = cfg
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
		return nil, errors.New("难度错误，仅支持Expert: exp, Master: mst, Append: apd，必须拼写正确")	
	}

	// 处理谱面图
	data, err := http.Get(p.pjskConfig.PJSK.Charts.RequestPath + id + level + ".png")
	if err != nil {
		return nil, err//ors.New("获取谱面图失败")
	}
	defer data.Body.Close()

	// 处理404
	if data.StatusCode == 404 {
		return nil, errors.New("未找到该谱面，请检查ID和难度是否正确")
	}

	chart, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, err//ors.New("读取谱面图失败")
	}
	//处理背景图
	bg, err := http.Get("https://sdvx.in/prsk/bg/" + id + "bg" + ".png")
	if err != nil {
		return nil, err//ors.New("获取背景图失败")
	}
	defer bg.Body.Close()
	background, err := io.ReadAll(bg.Body)
	if err != nil {
		return nil, err//ors.New("读取背景图失败")
	}

	// 合并两张图
	chartImage, err := png.Decode(bytes.NewReader(chart))
	if err != nil {
		return nil, err//ors.New("解析谱面图失败")
	}
	backgroundImage, err := png.Decode(bytes.NewReader(background))
	if err != nil {
		return nil, err//ors.New("解析背景图失败")
	}
	// 先画背景再画前景
	mergeImage := image.NewRGBA(chartImage.Bounds())
	draw.Draw(mergeImage, backgroundImage.Bounds(), backgroundImage, image.Point{}, draw.Src)
	draw.Draw(mergeImage, chartImage.Bounds(), chartImage, image.Point{}, draw.Over)
	// 把透明像素转为黑色
	transparentToBlack(mergeImage)

	// 编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, mergeImage)
	if err != nil {
		return nil, err//ors.New("合并图片失败")
	}
	mergeBytes := buf.Bytes()


	// 将图片保存到本地
	err = os.WriteFile(path, mergeBytes, 0644)
	if err != nil {
		return nil, errors.New("保存图片失败")
	}

	return mergeBytes, nil
}

// 检查难度是否合法
func checkLevel(level string) bool {
	switch level {
	case "exp":
		return true
	case "mst":
		return true
	case "apd":
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
	fileName := id + ".png"
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
	data, err := http.Get(p.pjskConfig.PJSK.Jackets.RequestPath + id + ".png")
	if err != nil {
		return nil, err//ors.New("获取封面图失败")
	}
	defer data.Body.Close()

	// 处理404
	if data.StatusCode == 404 {
		return nil, errors.New("未找到该封面，请检查ID是否正确")
	}
	jacket, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, err//ors.New("读取封面图失败")
	}

	// 将图片保存到本地
	err = os.WriteFile(path, jacket, 0644)
	if err != nil {
		return nil, errors.New("保存图片失败")
	}
	return jacket, nil
}