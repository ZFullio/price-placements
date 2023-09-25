package avito

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/zfullio/price-placements/v2/transport"
	"github.com/zfullio/price-placements/v2/validation"
	"io"
	"log"
	"net/http"
	"time"
)

type Feed struct {
	client       *http.Client
	url          string
	isGet        bool
	LastModified time.Time
	Data         Data
}

type Data struct {
	XMLName       xml.Name `xml:"Ads"`
	FormatVersion int      `xml:"formatVersion,attr"`
	Target        string   `xml:"target,attr"`
	Ad            []Ad     `xml:"Ad"`
}

func NewFeed(url string) *Feed {
	return &Feed{url: url}
}

type Ad struct {
	ID              string  `xml:"Id"`
	AdStatus        string  `xml:"AdStatus"`
	AllowEmail      string  `xml:"AllowEmail"`
	ContactPhone    string  `xml:"ContactPhone"`
	Latitude        string  `xml:"Latitude"`
	Longitude       string  `xml:"Longitude"`
	Description     string  `xml:"Description"`
	Category        string  `xml:"Category"`
	OperationType   string  `xml:"OperationType"`
	Price           int64   `xml:"Price"`
	Rooms           string  `xml:"Rooms"`
	Square          float32 `xml:"Square"`
	BalconyOrLoggia string  `xml:"BalconyOrLoggia"`
	KitchenSpace    float32 `xml:"KitchenSpace"`
	ViewFromWindows string  `xml:"ViewFromWindows"`
	CeilingHeight   string  `xml:"CeilingHeight"`
	LivingSpace     float32 `xml:"LivingSpace"`
	Decoration      string  `xml:"Decoration"`
	DealType        string  `xml:"DealType"`
	RoomType        struct {
		Option string `xml:"Option"`
	} `xml:"RoomType"`
	Status           string `xml:"Status"`
	Floor            int64  `xml:"Floor"`
	Floors           int64  `xml:"Floors"`
	HouseType        string `xml:"HouseType"`
	MarketType       string `xml:"MarketType"`
	PropertyRights   string `xml:"PropertyRights"`
	NewDevelopmentID string `xml:"NewDevelopmentId"`
	Images           struct {
		Image []struct {
			URL string `xml:"url,attr"`
		} `xml:"Image"`
	} `xml:"Images"`
}

func (f *Feed) Get(ctx context.Context) error {
	err := f.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("can't get feed info. Error:%w", err)
	}

	resp, err := transport.GetResponse(ctx, f.client, f.url)
	if err != nil {
		return fmt.Errorf("can't get feed data. Error:%w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(responseBody, &f.Data)
	if err != nil {
		return err
	}

	f.isGet = true

	//TODO Исправить значение f.LastModified

	return nil
}

func (f *Feed) GetInfo(ctx context.Context) error {
	resp, err := transport.GetOnlyHeader(ctx, f.client, f.url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	attributeLastModified := resp.Header.Get("Last-Modified")
	if attributeLastModified != "" {
		lastModifiedDate, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
		if err != nil {
			return err
		}

		f.LastModified = lastModifiedDate
	} else {
		log.Println("Header not contains `Last-Modified`")
	}

	return nil
}

func (f *Feed) Check() ([]string, error) {
	if !f.isGet {
		return nil, errors.New("feed not got")
	}

	results := make([]string, 0)

	if len(f.Data.Ad) < 2 {
		results = append(results, validation.MsgEmptyFeed)

		return results, nil
	}

	if len(f.Data.Ad) <= 10 {
		results = append(results, fmt.Sprintf("feed contains only %v items", len(f.Data.Ad)))

		return results, nil
	}

	for idx, lot := range f.Data.Ad {
		validation.CheckStringWithPos(idx, "Ad", "ID", lot.ID, &results)
		id := lot.ID
		validation.CheckStringWithID(id, "Ad", "ContactPhone", lot.ContactPhone, &results)
		validation.CheckStringWithID(id, "Ad", "Description", lot.Description, &results)
		validation.CheckStringWithID(id, "Ad", "Category", lot.Category, &results)
		validation.CheckZeroWithID(id, "Ad", "Price", int(lot.Price), &results)
		validation.CheckStringWithID(id, "Ad", "OperationType", lot.OperationType, &results)
		validation.CheckStringWithID(id, "Ad", "MarketType", lot.MarketType, &results)
		validation.CheckStringWithID(id, "Ad", "HouseType", lot.HouseType, &results)
		validation.CheckZeroWithID(id, "Ad", "Floor", int(lot.Floor), &results)
		validation.CheckZeroWithID(id, "Ad", "Floors", int(lot.Floors), &results)
		validation.CheckStringWithID(id, "Ad", "Rooms", lot.Rooms, &results)
		validation.CheckZeroWithID(id, "Ad", "Square", lot.Square, &results)

		if lot.LivingSpace == 0 && lot.Rooms != "Студия" {
			results = append(results, fmt.Sprintf("field LivingSpace is empty. InternalID: %v", lot.ID))
		}

		validation.CheckStringWithID(id, "Ad", "Status", lot.Status, &results)
		validation.CheckStringWithID(id, "Ad", "NewDevelopmentId", lot.NewDevelopmentID, &results)
		validation.CheckStringWithID(id, "Ad", "PropertyRights", lot.PropertyRights, &results)
		validation.CheckStringWithID(id, "Ad", "Decoration", lot.Decoration, &results)

		if lot.Floor > lot.Floors {
			results = append(results, fmt.Sprintf("field Floor is bigger than Floors. InternalID: %v", lot.ID))
		}

		for idx, image := range lot.Images.Image {
			validation.CheckStringWithPos(idx, "Images.Image", "URL", image.URL, &results)
		}

		if len(lot.Images.Image) < 3 || len(lot.Images.Image) > 40 {
			results = append(results, fmt.Sprintf("field Images.Image contains '%v' items. InternalID: %v", len(lot.Images.Image), lot.ID))
		}
	}

	return results, nil
}

type Developments struct {
	Region []Region `xml:"Region"`
}

type Region struct {
	Name string `xml:"name,attr"`
	City []City `xml:"City"`
}

type City struct {
	Name   string   `xml:"name,attr"`
	Object []Object `xml:"Object"`
}

type Object struct {
	ID        string  `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	Address   string  `xml:"address,attr"`
	Developer string  `xml:"developer,attr"`
	Housing   []House `xml:"Housing"`
}

type House struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Address string `xml:"address,attr"`
}

func (f *Feed) GetDevelopments(ctx context.Context) (Developments, error) {
	url := "https://autoload.avito.ru/format/New_developments.xml"

	resp, err := transport.GetResponse(ctx, f.client, url)
	if err != nil {
		return Developments{}, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Developments{}, err
	}

	developments := Developments{}
	err = xml.Unmarshal(responseBody, &developments)
	if err != nil {
		return Developments{}, err
	}

	return developments, err
}
