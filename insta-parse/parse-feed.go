package insta_parse

import (
	"errors"
	"fmt"
	"github.com/Davincible/goinsta/v3"
	"time"
)

func GetData(item *goinsta.Item) (mediaType []int, url []string, text string, err error) {
	if item == nil {
		return nil, nil, "", errors.New("item is empty")
	}

	switch item.MediaType {
	case 1:
		mediaType := []int{item.MediaType}
		url := make([]string, 0, 1)
		url = append(url, goinsta.GetBest(item.Images.Versions))
		text := getBlogName(item) + "\n" + item.Caption.Text
		return mediaType, url, text, nil
	case 2:
		mediaType := []int{item.MediaType}
		url := make([]string, 0, 1)
		url = append(url, goinsta.GetBest(item.Videos))
		text := getBlogName(item) + "\n" + item.Caption.Text
		return mediaType, url, text, nil
	case 8:
		url := make([]string, 0, len(item.CarouselMedia))
		mediaType := make([]int, 0, len(item.CarouselMedia))
		for _, it := range item.CarouselMedia {
			t, u, _, err := GetData(&it)
			if err != nil {
				continue
			}
			mediaType = append(mediaType, t[0])
			url = append(url, u[0])
			time.Sleep(100 * time.Millisecond)
		}
		text := getBlogName(item) + "\n" + item.Caption.Text
		return mediaType, url, text, nil

	default:
		return nil, nil, "", errors.New("unknown MediaType")
	}
}

func getBlogName(item *goinsta.Item) string {
	return fmt.Sprintf("%s (@%s)", item.User.FullName, item.User.Username)
}
