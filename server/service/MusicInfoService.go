package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type SearchOptions struct {
	Title    string
	Author   string
	MinLevel int
	MaxLevel int
	PageNo   int
	PageSize int
}

type MusicCategory struct {
	MusicCategoryName string `json:"musicCategoryName"`
}

type MusicCategories []MusicCategory

func (m *MusicCategories) UnmarshalJSON(data []byte) error {
	// categories有两种格式
	//[{ "musicCategoryName": "mv" }]
	//["mv"]
	//一种是字典，另一种直接是字符串
	//都要可以解析
	var asObjects []MusicCategory
	if err := json.Unmarshal(data, &asObjects); err == nil {
		*m = asObjects
		return nil
	}

	var asStrings []string
	if err := json.Unmarshal(data, &asStrings); err == nil {
		categories := make([]MusicCategory, 0, len(asStrings))
		for _, item := range asStrings {
			categories = append(categories, MusicCategory{MusicCategoryName: item})
		}
		*m = categories
		return nil
	}

	return errors.New("invalid categories format")
}

type DifficultyInfo struct {
	PlayLevel          int `json:"playLevel"`
	ReleaseConditionID int `json:"releaseConditionId"`
	TotalNoteCount     int `json:"totalNoteCount"`
}

type MusicInfo struct {
	ID                 int                       `json:"id"`
	Seq                int                       `json:"seq"`
	ReleaseConditionID int                       `json:"releaseConditionId"`
	Categories         MusicCategories           `json:"categories"`
	Title              string                    `json:"title"`
	ChineseTitle       string                    `json:"chinese_title"`
	Pronunciation      string                    `json:"pronunciation"`
	CreatorArtistID    int                       `json:"creatorArtistId"`
	Lyricist           string                    `json:"lyricist"`
	Composer           string                    `json:"composer"`
	Arranger           string                    `json:"arranger"`
	DancerCount        int                       `json:"dancerCount"`
	SelfDancerPosition int                       `json:"selfDancerPosition"`
	AssetbundleName    string                    `json:"assetbundleName"`
	PublishedAt        int64                     `json:"publishedAt"`
	ReleasedAt         int64                     `json:"releasedAt"`
	FillerSec          float64                   `json:"fillerSec"`
	Infos              []map[string]interface{}  `json:"infos"`
	IsNewlyWritten     bool                      `json:"isNewlyWrittenMusic"`
	IsFullLength       bool                      `json:"isFullLength"`
	Difficulties       map[string]DifficultyInfo `json:"difficulties"`
}

type musicDifficultyRecord struct {
	MusicID            int    `json:"musicId"`
	MusicDifficulty    string `json:"musicDifficulty"`
	PlayLevel          int    `json:"playLevel"`
	ReleaseConditionID int    `json:"releaseConditionId"`
	TotalNoteCount     int    `json:"totalNoteCount"`
}

type bulkResponse struct {
	Errors bool `json:"errors"`
}

type searchResult struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source MusicInfo `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (p *PJSKService) UpdateMusicInfos() (int, error) {
	var infos []MusicInfo
	if err := p.fetchJSON(p.pjskConfig.PJSK.Infos.RequestPath, &infos); err != nil {
		return 0, err
	}

	var difficults []musicDifficultyRecord
	if err := p.fetchJSON(p.pjskConfig.PJSK.Difficulties.RequestPath, &difficults); err != nil {
		return 0, err
	}

	chineseTitles, err := loadChineseTitleMap("resources/chinese_title.json")
	if err != nil {
		return 0, err
	}

	difficultMap := make(map[int]map[string]DifficultyInfo)
	for _, item := range difficults {
		if difficultMap[item.MusicID] == nil {
			difficultMap[item.MusicID] = make(map[string]DifficultyInfo)
		}
		difficultMap[item.MusicID][item.MusicDifficulty] = DifficultyInfo{
			PlayLevel:          item.PlayLevel,
			ReleaseConditionID: item.ReleaseConditionID,
			TotalNoteCount:     item.TotalNoteCount,
		}
	}

	for i := range infos {
		key := fmt.Sprintf("%d", infos[i].ID)
		infos[i].ChineseTitle = chineseTitles[key]
		if infos[i].Difficulties == nil {
			infos[i].Difficulties = make(map[string]DifficultyInfo)
		}
		if item, ok := difficultMap[infos[i].ID]; ok {
			infos[i].Difficulties = item
		}
	}

	if err := p.recreateMusicIndex(); err != nil {
		return 0, err
	}
	if err := p.bulkIndexMusicInfos(infos); err != nil {
		return 0, err
	}
	return len(infos), nil
}

func (p *PJSKService) SearchMusicInfos(options SearchOptions) ([]MusicInfo, int64, error) {
	must := make([]map[string]interface{}, 0)
	//o对标题的查询需要构建比较完备的打分机制，完整查询词稳定靠前
	//match_phrase(chinese_title)，boost=20
	//match_phrase(title)，boost=15
	//match(... operator=and) 中权重
	//普通 match 低权重兜底
	if options.Title != "" {
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match_phrase": map[string]interface{}{
							"chinese_title": map[string]interface{}{
								"query": options.Title,
								"boost": 20,
							},
						},
					},
					{
						"match_phrase": map[string]interface{}{
							"title": map[string]interface{}{
								"query": options.Title,
								"boost": 15,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"chinese_title": map[string]interface{}{
								"query":    options.Title,
								"operator": "and",
								"boost":    8,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query":    options.Title,
								"operator": "and",
								"boost":    6,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"chinese_title": map[string]interface{}{
								"query": options.Title,
								"boost": 2,
							},
						},
					},
					{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query": options.Title,
								"boost": 1,
							},
						},
					},
				},
				"minimum_should_match": 1,
			},
		})
	}
	if options.Author != "" {
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"match": map[string]string{"lyricist": options.Author}},
					{"match": map[string]string{"composer": options.Author}},
					{"match": map[string]string{"arranger": options.Author}},
				},
				"minimum_should_match": 1,
			},
		})
	}
	if levelFilter := buildLevelFilter(options.MinLevel, options.MaxLevel); levelFilter != nil {
		must = append(must, levelFilter)
	}

	query := map[string]interface{}{
		"track_total_hits": true,
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"id": map[string]string{"order": "asc"}},
		},
	}
	if len(must) == 0 {
		query["query"] = map[string]interface{}{"match_all": map[string]interface{}{}}
	} else {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		}
	}

	if options.PageNo > 0 && options.PageSize > 0 {
		query["from"] = (options.PageNo - 1) * options.PageSize
		query["size"] = options.PageSize
	} else {
		query["size"] = 10000
	}

	payload, err := json.Marshal(query)
	if err != nil {
		return nil, 0, err
	}

	resp, err := p.esRequest(http.MethodGet, fmt.Sprintf("/%s/_search", p.getMusicIndexName()), bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("es search failed: %s", string(body))
	}

	var result searchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}
	records := make([]MusicInfo, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		records = append(records, hit.Source)
	}
	return records, result.Hits.Total.Value, nil
}

func buildLevelFilter(minLevel int, maxLevel int) map[string]interface{} {
	if minLevel <= 0 && maxLevel <= 0 {
		return nil
	}

	rangeBody := make(map[string]int)
	if minLevel > 0 {
		rangeBody["gte"] = minLevel
	}
	if maxLevel > 0 {
		rangeBody["lte"] = maxLevel
	}

	fields := []string{
		"difficulties.easy.playLevel",
		"difficulties.normal.playLevel",
		"difficulties.hard.playLevel",
		"difficulties.expert.playLevel",
		"difficulties.master.playLevel",
		"difficulties.append.playLevel",
	}
	should := make([]map[string]interface{}, 0, len(fields))
	for _, field := range fields {
		should = append(should, map[string]interface{}{
			"range": map[string]interface{}{
				field: rangeBody,
			},
		})
	}
	return map[string]interface{}{
		"bool": map[string]interface{}{
			"should":               should,
			"minimum_should_match": 1,
		},
	}
}

// 直接请求本地了，这玩意服务器上几年没更新了，不如我自己加
func loadChineseTitleMap(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var titles map[string]string
	if err := json.Unmarshal(content, &titles); err != nil {
		return nil, err
	}
	return titles, nil
}

// 获取ES索取名字，默认是pjsk_music_infos
func (p *PJSKService) getMusicIndexName() string {
	if p.pjskConfig.Elasticsearch.Index == "" {
		return "pjsk_music_infos"
	}
	return p.pjskConfig.Elasticsearch.Index
}

func (p *PJSKService) esRequest(method string, uri string, body io.Reader) (*http.Response, error) {
	address := strings.TrimSpace(p.pjskConfig.Elasticsearch.Address)
	if address == "" {
		return nil, errors.New("elasticsearch address is empty")
	}
	address = strings.TrimRight(address, "/")

	req, err := http.NewRequest(method, address+uri, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	return client.Do(req)
}

func (p *PJSKService) recreateMusicIndex() error {
	index := p.getMusicIndexName()

	deleteResp, err := p.esRequest(http.MethodDelete, "/"+index, nil)
	if err != nil {
		return err
	}
	deleteResp.Body.Close()

	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":            map[string]string{"type": "integer"},
				"title":         map[string]string{"type": "text"},
				"chinese_title": map[string]string{"type": "text"},
				"lyricist":      map[string]string{"type": "text"},
				"composer":      map[string]string{"type": "text"},
				"arranger":      map[string]string{"type": "text"},
				"difficulties": map[string]interface{}{
					"properties": map[string]interface{}{
						"easy":   map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
						"normal": map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
						"hard":   map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
						"expert": map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
						"master": map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
						"append": map[string]interface{}{"properties": map[string]interface{}{"playLevel": map[string]string{"type": "integer"}}},
					},
				},
			},
		},
	}
	payload, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	createResp, err := p.esRequest(http.MethodPut, "/"+index, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer createResp.Body.Close()
	if createResp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create es index failed: %s", string(body))
	}
	return nil
}

func (p *PJSKService) bulkIndexMusicInfos(infos []MusicInfo) error {
	var builder strings.Builder
	writer := bufio.NewWriter(&builder)
	index := p.getMusicIndexName()
	for _, info := range infos {
		meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%d"}}`, index, info.ID)
		if _, err := writer.WriteString(meta + "\n"); err != nil {
			return err
		}
		body, err := json.Marshal(info)
		if err != nil {
			return err
		}
		if _, err := writer.Write(body); err != nil {
			return err
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	resp, err := p.esRequest(http.MethodPost, "/_bulk", strings.NewReader(builder.String()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk index failed: %s", string(body))
	}

	var result bulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Errors {
		return errors.New("bulk index contains failed items")
	}
	return nil
}

// fetchJSON 获取服务器的JSON内容
func (p *PJSKService) fetchJSON(url string, target interface{}) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s", string(body))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
