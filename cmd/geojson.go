package main

import "github.com/losinggeneration/geojson"

type geojsoner interface {
	GeoJSON() interface{}
}

func GetGeoJSON(g GopherCon) interface{} {
	return geojson.GeoJSON{
		Feature: &geojson.Feature{
			Geometry: &geojson.Geometry{
				Point: &geojson.Point{
					Coordinates: geojson.Position{
						39.742329, -104.9965061,
					},
				},
			},
			Properties: geojson.Properties{
				"id":   g.Id,
				"name": g.Name,
			},
		},
	}
}
