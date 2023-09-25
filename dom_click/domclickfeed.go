package domclick

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
	XMLName xml.Name `xml:"complexes"`
	Complex struct {
		ID        string `xml:"id"`
		Name      string `xml:"name"`
		Latitude  string `xml:"latitude"`
		Longitude string `xml:"longitude"`
		Address   string `xml:"address"`
		Images    struct {
			Image []string `xml:"image"`
		} `xml:"images"`
		DescriptionMain struct {
			Title string `xml:"title"`
			Text  string `xml:"text"`
		} `xml:"description_main"`
		Infrastructure struct {
			Parking      string `xml:"parking"`
			Security     string `xml:"security"`
			FencedArea   string `xml:"fenced_area"`
			SportsGround string `xml:"sports_ground"`
			Playground   string `xml:"playground"`
			School       string `xml:"school"`
			Kindergarten string `xml:"kindergarten"`
		} `xml:"infrastructure"`
		ProfitsMain struct {
			ProfitMain []struct {
				Title string `xml:"title"`
				Text  string `xml:"text"`
				Image string `xml:"image"`
			} `xml:"profit_main"`
		} `xml:"profits_main"`
		ProfitsSecondary struct {
			ProfitSecondary []struct {
				Title string `xml:"title"`
				Text  string `xml:"text"`
				Image string `xml:"image"`
			} `xml:"profit_secondary"`
		} `xml:"profits_secondary"`
		Buildings struct {
			Building []struct {
				ID            string `xml:"id"`
				Fz214         string `xml:"fz_214"`
				Name          string `xml:"name"`
				Floors        int64  `xml:"floors"`
				BuildingState string `xml:"building_state"`
				BuiltYear     int64  `xml:"built_year"`
				ReadyQuarter  int64  `xml:"ready_quarter"`
				BuildingType  string `xml:"building_type"`
				Image         string `xml:"image"`
				Flats         struct {
					Flat []Flat `xml:"flat"`
				} `xml:"flats"`
			} `xml:"building"`
		} `xml:"buildings"`
		SalesInfo struct {
			SalesPhone              string `xml:"sales_phone"`
			ResponsibleOfficerPhone string `xml:"responsible_officer_phone"`
			SalesAddress            string `xml:"sales_address"`
			SalesLatitude           string `xml:"sales_latitude"`
			SalesLongitude          string `xml:"sales_longitude"`
			Timezone                string `xml:"timezone"`
			WorkDays                struct {
				WorkDay []struct {
					Day     string `xml:"day"`
					OpenAt  string `xml:"open_at"`
					CloseAt string `xml:"close_at"`
				} `xml:"work_day"`
			} `xml:"work_days"`
		} `xml:"sales_info"`
		Developer struct {
			ID    string `xml:"id"`
			Name  string `xml:"name"`
			Phone string `xml:"phone"`
			Site  string `xml:"site"`
			Logo  string `xml:"logo"`
		} `xml:"developer"`
	} `xml:"complex"`
}

type Flat struct {
	FlatID      string  `xml:"flat_id"`
	Apartment   string  `xml:"apartment"`
	Floor       int64   `xml:"floor"`
	Room        *int64  `xml:"room"`
	Plan        string  `xml:"plan"`
	Balcony     string  `xml:"balcony"`
	Renovation  string  `xml:"renovation"`
	Price       float32 `xml:"price"`
	Area        float32 `xml:"area"`
	LivingArea  float32 `xml:"living_area"`
	KitchenArea float32 `xml:"kitchen_area"`
	RoomsArea   struct {
		Area []string `xml:"area"`
	} `xml:"rooms_area"`
	Bathroom     string `xml:"bathroom"`
	HousingType  string `xml:"housing_type"`
	Decoration   int64  `xml:"decoration"`
	ReadyHousing string `xml:"ready_housing"`
}

func NewFeed(url string) *Feed {
	return &Feed{url: url}
}

func (f *Feed) Get(ctx context.Context) error {
	resp, err := transport.GetResponse(ctx, f.client, f.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	AttributeLastModified := resp.Header.Get(transport.HeaderLastModified)
	if AttributeLastModified != "" {
		lastModifiedDate, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
		if err != nil {
			return err
		}

		f.LastModified = lastModifiedDate
	} else {
		log.Println("Header not contains `Last-Modified`")
	}

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
	residence := &f.Data.Complex
	if len(residence.Buildings.Building) < 2 {
		results = append(results, validation.MsgEmptyFeed)

		return results, nil
	}

	path := "Complex"

	validation.CheckString(path, "ID", residence.ID, &results)
	validation.CheckString(path, "Name", residence.Name, &results)
	validation.CheckString(path, "Address", residence.Address, &results)
	validation.CheckString(path, "Latitude", residence.Latitude, &results)
	validation.CheckString(path, "Longitude", residence.Longitude, &results)

	for idx, image := range residence.Images.Image {
		validation.CheckStringWithPos(idx, "Complex.Images.Image", "Image", image, &results)
	}

	path = "Complex.DescriptionMain"
	descriptionMain := &residence.DescriptionMain
	validation.CheckString(path, "Title", descriptionMain.Title, &results)
	validation.CheckString(path, "Text", descriptionMain.Text, &results)

	path = "Complex.ProfitsMain.ProfitMain"

	profits := residence.ProfitsMain.ProfitMain
	for idx, profit := range profits {
		validation.CheckStringWithPos(idx, path, "Title", profit.Title, &results)
		validation.CheckStringWithPos(idx, path, "Text", profit.Text, &results)
		validation.CheckStringWithPos(idx, path, "Image", profit.Image, &results)
	}

	path = "Complex.Buildings.Building"

	buildings := f.Data.Complex.Buildings.Building
	for pos, building := range buildings {
		validation.CheckStringWithPos(pos, path, "ID", building.ID, &results)
		validation.CheckStringWithID(building.ID, path, "Fz214", building.Fz214, &results)
		validation.CheckStringWithID(building.ID, path, "Name", building.Name, &results)
		validation.CheckZeroWithID(building.ID, path, "Floors", int(building.Floors), &results)
		validation.CheckStringWithID(building.ID, path, "BuildingState", building.BuildingState, &results)
		validation.CheckZeroWithID(building.ID, path, "BuiltYear", int(building.BuiltYear), &results)
		validation.CheckZeroWithID(building.ID, path, "ReadyQuarter", int(building.ReadyQuarter), &results)
		validation.CheckStringWithID(building.ID, path, "BuildingType", building.BuildingType, &results)

		if building.BuiltYear < int64(time.Now().Year()) && building.BuildingState == "unfinished" {
			results = append(results, fmt.Sprintf("BuildingState == unfinished for %v. InternalID: %v", building.BuiltYear, building.ID))
		}

		f.checkLots(building.Flats.Flat, int(building.Floors), &results)
	}

	path = "Complex.SalesInfo"
	salesInfo := &f.Data.Complex.SalesInfo
	validation.CheckString(path, "SalesPhone", salesInfo.SalesPhone, &results)
	validation.CheckString(path, "SalesAddress", salesInfo.SalesAddress, &results)
	validation.CheckString(path, "SalesLatitude", salesInfo.SalesLatitude, &results)
	validation.CheckString(path, "SalesLongitude", salesInfo.SalesLongitude, &results)

	path = "Complex.Developer"
	developer := &f.Data.Complex.Developer
	validation.CheckString(path, "Name", developer.Name, &results)
	validation.CheckString(path, "Phone", developer.Phone, &results)
	validation.CheckString(path, "Site", developer.Site, &results)
	validation.CheckString(path, "Logo", developer.Logo, &results)

	return results, nil
}

func (f *Feed) checkLots(flats []Flat, floors int, results *[]string) {
	path := "Flats.Flat"
	for idx, lot := range flats {
		validation.CheckStringWithPos(idx, path, "FlatID", lot.FlatID, results)
		validation.CheckZeroWithID(lot.FlatID, path, "Floor", int(lot.Floor), results)

		if lot.Room == nil {
			*results = append(*results, fmt.Sprintf("Field Flats.Room is empty. InternalID: %v", lot.FlatID))
		}

		validation.CheckStringWithID(lot.FlatID, path, "Plan", lot.Plan, results)
		validation.CheckStringWithID(lot.FlatID, path, "Balcony", lot.Balcony, results)
		validation.CheckZeroWithID(lot.FlatID, path, "Price", lot.Price, results)
		validation.CheckZeroWithID(lot.FlatID, path, "Area", lot.Area, results)

		isOk := validation.CheckZeroWithID(lot.FlatID, path, "LivingArea", lot.LivingArea, results)
		if !isOk {
			for i, room := range lot.RoomsArea.Area {
				if room == "" {
					*results = append(*results, fmt.Sprintf("Field Flats.Flat.RoomsArea.Area[%v] is empty. InternalID: %v", i, lot.FlatID))
				}
			}
		}

		validation.CheckZeroWithID(lot.FlatID, path, "KitchenArea", lot.KitchenArea, results)
		validation.CheckStringWithID(lot.FlatID, path, "Bathroom", lot.Bathroom, results)

		if lot.Floor > int64(floors) {
			*results = append(*results, fmt.Sprintf("Field Flats.Flat.Floor is bigger than building.Floors. InternalID: %v", lot.FlatID))
		}
	}
}
