package cian

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/zfullio/price-placements/transport"
	"github.com/zfullio/price-placements/validation"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	FeedVersion string   `xml:"feed_version"`
	Object      []Object `xml:"object"`
}

func NewFeed(url string) *Feed {
	return &Feed{url: url}
}

type Object struct {
	ExternalId  string `xml:"ExternalId"`
	Description string `xml:"Description"`
	Address     string `xml:"Address"`
	Coordinates struct {
		Lat float32 `xml:"Lat"`
		Lng float32 `xml:"Lng"`
	} `xml:"Coordinates"`
	CadastralNumber string `xml:"CadastralNumber"`
	Phones          struct {
		PhoneSchema struct {
			CountryCode string `xml:"CountryCode"`
			Number      string `xml:"Number"`
		} `xml:"PhoneSchema"`
	} `xml:"Phones"`
	LayoutPhoto struct {
		IsDefault bool   `xml:"IsDefault"`
		FullUrl   string `xml:"FullUrl"`
	} `xml:"LayoutPhoto"`
	Photos struct {
		PhotoSchema []PhotoSchema `xml:"PhotoSchema"`
	} `xml:"Photos"`
	Category              string  `xml:"Category"`
	RoomType              string  `xml:"RoomType"`
	FlatRoomsCount        int64   `xml:"FlatRoomsCount"`
	TotalArea             float32 `xml:"TotalArea"`
	LivingArea            float32 `xml:"LivingArea"`
	KitchenArea           float32 `xml:"KitchenArea"`
	ProjectDeclarationUrl string  `xml:"ProjectDeclarationUrl"`
	FloorNumber           int64   `xml:"FloorNumber"`
	CombinedWcsCount      int64   `xml:"CombinedWcsCount"`
	Building              struct {
		FloorsCount         int64  `xml:"FloorsCount"`
		MaterialType        string `xml:"MaterialType"`
		PassengerLiftsCount int64  `xml:"PassengerLiftsCount"`
		CargoLiftsCount     int64  `xml:"CargoLiftsCount"`
		Parking             struct {
			Type string `xml:"Type"`
		} `xml:"Parking"`
		Deadline struct {
			Quarter    string `xml:"Quarter"`
			Year       int64  `xml:"Year"`
			IsComplete bool   `xml:"IsComplete"`
		} `xml:"Deadline"`
	} `xml:"Building"`
	BargainTerms struct {
		Price           CustomFloat64 `xml:"Price"`
		Currency        string        `xml:"Currency"`
		MortgageAllowed bool          `xml:"MortgageAllowed"`
		SaleType        string        `xml:"SaleType"`
	} `xml:"BargainTerms"`
	JKSchema struct {
		ID    int32  `xml:"Id"`
		Name  string `xml:"Name"`
		House struct {
			ID   int32  `xml:"Id"`
			Name string `xml:"Name"`
			Flat struct {
				FlatNumber    string `xml:"FlatNumber"`
				SectionNumber string `xml:"SectionNumber"`
				FlatType      string `xml:"FlatType"`
			} `xml:"Flat"`
		} `xml:"House"`
	} `xml:"JKSchema"`
	Decoration      string  `xml:"Decoration"`
	WindowsViewType string  `xml:"WindowsViewType"`
	CeilingHeight   float32 `xml:"CeilingHeight"`
	Undergrounds    struct {
		UndergroundInfoSchema []struct {
			TransportType string `xml:"TransportType"`
			Time          int64  `xml:"Time"`
			ID            int64  `xml:"Id"`
		} `xml:"UndergroundInfoSchema"`
	} `xml:"Undergrounds"`
	IsApartments bool `xml:"isApartments"`
}

type PhotoSchema struct {
	FullUrl   string `xml:"FullUrl"`
	IsDefault bool   `xml:"IsDefault"`
}

type CustomFloat64 struct {
	Float64 float64
}

func (cf *CustomFloat64) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	s = strings.ReplaceAll(s, ",", ".")
	float, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	cf.Float64 = float
	return nil
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

	err = xml.Unmarshal(responseBody, &f)
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

	attributeLastModified := resp.Header.Get(transport.HeaderLastModified)
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

	if len(f.Data.Object) < 2 {
		results := append(results, validation.MsgEmptyFeed)
		return results, nil
	}

	if len(f.Data.Object) <= 10 {
		results = append(results, fmt.Sprintf("feed contains only %v items", len(f.Data.Object)))
		return results, nil
	}
	for idx, lot := range f.Data.Object {
		id := lot.ExternalId

		if lot.ExternalId == "" {
			results = append(results, fmt.Sprintf("field ExternalId is empty. Position: %v", idx))
		}
		validation.CheckStringWithID(id, "object", "Address", lot.Address, &results)
		validation.CheckStringWithID(id, "object.Phones.PhoneSchema", "CountryCode", lot.Phones.PhoneSchema.CountryCode, &results)
		validation.CheckStringWithID(id, "object.Phones.PhoneSchema", "Number", lot.Phones.PhoneSchema.Number, &results)
		validation.CheckStringWithID(id, "object.LayoutPhoto.FullUrl", "IsDefault", lot.LayoutPhoto.FullUrl, &results)
		validation.CheckStringWithID(id, "object", "Category", lot.Category, &results)

		for idx, photoSchema := range lot.Photos.PhotoSchema {
			validation.CheckStringWithPos(idx, "object.Photos.PhotoSchema", "FullUrl", photoSchema.FullUrl, &results)
		}

		validation.CheckZeroWithID(id, "object", "FlatRoomsCount", int(lot.FlatRoomsCount), &results)
		validation.CheckZeroWithID(id, "object", "TotalArea", int(lot.TotalArea), &results)
		validation.CheckZeroWithID(id, "object", "FloorNumber", int(lot.FloorNumber), &results)
		validation.CheckZeroWithID(id, "object.Building", "FloorsCount", int(lot.Building.FloorsCount), &results)
		validation.CheckZeroWithID(id, "object.Building.Deadline", "Year", int(lot.Building.Deadline.Year), &results)
		validation.CheckStringWithID(id, "object.Building.Deadline", "Quarter", lot.Building.Deadline.Quarter, &results)
		validation.CheckZeroWithID(id, "object.BargainTerms.Price", "Price", int(lot.BargainTerms.Price.Float64), &results)
		validation.CheckZeroWithID(id, "object.JKSchema", "Id", int(lot.JKSchema.ID), &results)
		validation.CheckStringWithID(id, "object.JKSchema", "Name", lot.JKSchema.Name, &results)
		validation.CheckZeroWithID(id, "object.JKSchema.House", "Id", int(lot.JKSchema.House.ID), &results)
		validation.CheckStringWithID(id, "object.JKSchema.House", "Name", lot.JKSchema.House.Name, &results)

		if lot.Building.Deadline.Year < int64(time.Now().Year()) && lot.Building.Deadline.IsComplete == false {
			results = append(results, fmt.Sprintf("field Building.Deadline is False for %v. InternalID: %v", lot.Building.Deadline.Year, lot.ExternalId))
		}
		if lot.FloorNumber > lot.Building.FloorsCount {
			results = append(results, fmt.Sprintf("field FloorNumber is greater than Building.FloorsCount. InternalID: %v", lot.ExternalId))
		}
		if len(lot.Photos.PhotoSchema) < 3 {
			results = append(results, fmt.Sprintf("field Photos.PhotoSchema contains '%v' items. InternalID: %v", len(lot.Photos.PhotoSchema), lot.ExternalId))
		}
	}

	return results, nil
}
