package commons

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FilterRequest struct {
    Sort     	string  			`bson:"sort" json:"sort"`
	IsAscending bool 				`bson:"is_ascending" json:"is_ascending"`
    Page     	int      			`bson:"page" json:"page"`
    PageSize 	int      			`bson:"page_size" json:"page_size"`
    Filters  	[]Filter 			`bson:"filters" json:"filters"`
}

type Filter struct {
	Operator string `bson:"operator" json:"operator"`
	Fields []Field `bson:"fields" json:"fields"`
}

type Field struct {
	Name 		string `bson:"name" json:"name"`
	Value 		string `bson:"value" json:"value"`
	Type 		string `bson:"type" json:"type"`
	MinValue 	string `bson:"min_value" json:"min_value"`
	MaxValue 	string `bson:"max_value" json:"max_value"`
}

func (filter *FilterRequest) GetFilter() bson.M {

	returnFilter := bson.M{}
	andConditions := []bson.M{}

	for _, filterItem := range filter.Filters {
		itemFilters := []bson.M{}
		for _, item := range filterItem.Fields {
			switch item.Type {
			case "string":
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$regex": item.Value, "$options": "i"}})			
			case "date":
                dateValue, err := time.Parse("2006-01-02", item.Value)
                if err != nil {
                    // handle error
                    continue
                }				
				startDate := time.Date(dateValue.Year(), dateValue.Month(), dateValue.Day(), 0, 0, 0, 0, dateValue.Location())
				endDate := time.Date(dateValue.Year(), dateValue.Month(), dateValue.Day(), 23, 59, 59, 999999999, dateValue.Location())
				itemFilters = append(itemFilters, bson.M{
					item.Name: bson.M{
						"$gte": primitive.NewDateTimeFromTime(startDate),
						"$lte":  primitive.NewDateTimeFromTime(endDate),
					},
				})				
			case "array_object":
				values := strings.Split(item.Value, ",")
				for _, value := range values {
					itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$regex": value, "$options": "i"}})
				}
			case "array_string":
				values := strings.Split(item.Value, ",")		
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$in": values}})			
			case "boolean", "checkbox":
				// convert string to boolean
				value, err := strconv.ParseBool(item.Value)
				if err != nil {
					value = false
				}
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$eq": value}})
			case "number":
				value, err := strconv.ParseInt(item.Value, 10, 64)
				if err != nil {
					value = 0
				}
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$eq": value}})
			case "number_range":				
				min_value, err := strconv.ParseInt(item.MinValue, 10, 64)				
				if err != nil {
					min_value = 0
				}				
				max_value, err := strconv.ParseInt(item.MaxValue, 10, 64)
				if err != nil {
					// Max value of int
					max_value = 9223372036854775807
				}
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$gte": min_value, "$lte": max_value}})
			default:
				itemFilters = append(itemFilters, bson.M{item.Name: bson.M{"$regex": item.Value, "$options": "i"}})
			}			
		}

		if len(itemFilters) > 0 {
			filterOperator := bson.M{"$" + filterItem.Operator: itemFilters}
			andConditions = append(andConditions, filterOperator)
		}
	}

	if len(andConditions) != 0 {
		returnFilter["$and"] = andConditions
	}

	filterJSON, err := json.MarshalIndent(returnFilter, "", "  ")
    if err != nil {
        log.Fatalf("Error converting filter to JSON: %v", err)
    }

    // Print the JSON string
    fmt.Println("MongoDB Query Filter:", string(filterJSON))
	
	return returnFilter
}