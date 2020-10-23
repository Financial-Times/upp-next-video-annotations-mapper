[![CircleCI](https://circleci.com/gh/Financial-Times/upp-next-video-annotations-mapper.svg?style=svg)](https://circleci.com/gh/Financial-Times/upp-next-video-annotations-mapper) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/upp-next-video-annotations-mapper)](https://goreportcard.com/report/github.com/Financial-Times/upp-next-video-annotations-mapper) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/upp-next-video-annotations-mapper/badge.svg)](https://coveralls.io/github/Financial-Times/upp-next-video-annotations-mapper)

# Next Video Annotations Mapper (next-video-annotations-mapper)

Next Video Annotations Mapper transforms the metadata as received from Next within the video content to an internal format acceptable for further processing and storing on Neo4J.
Video content from Next is got from the Kafka(-bridge) queue where the application listens on topic NativeCmsPublicationEvents, the annotations and video uuid information is
used for transformation, resulting content being put back to Kafka(-bridge) on ConceptAnnotations topic.

## Installation

In order to install, execute the following steps:

```shell script
go get -u github.com/Financial-Times/upp-next-video-annotations-mapper
cd $GOPATH/src/github.com/Financial-Times/upp-next-video-annotations-mapper
go build -mod=readonly .
```

## Running

Locally with default configuration:

```
go install
$GOPATH/bin/upp-next-video-annotations-mapper
```

Locally with properties set:

```
go install
$GOPATH/bin/upp-next-video-annotations-mapper \
--app_port "8080" \
--q_addr "http://localhost:8080" \
--q_group "videoAnnotationsMapper" \
--q_read_topic "NativeCmsPublicationEvents" \
--q_read_queue "kafka" \
--q_write_topic "V1ConceptAnnotations" \
--q_write_queue "kafka" \
```

With Docker:

`docker build -t coco/upp-next-video-annotations-mapper .`

`docker run -ti coco/upp-next-video-annotations-mapper`

```
docker run -ti
--env "APP_PORT=8080" \
--env "Q_ADDR=http://localhost:8080" \
--env "Q_GROUP=videoAnnotationsMapper" \
--env "Q_READ_TOPIC=NativeCmsPublicationEvents" \
--env "Q_READ_QUEUE=kafka" \
--env "Q_WRITE_TOPIC=V1ConceptAnnotations" \
--env "Q_WRITE_QUEUE=kafka" \
coco/upp-next-video-annotations-mapper
```

When deployed locally arguments are optional.

## Endpoints
### POST
/map

This is a verification operation as it is not used on processing flow.

Example
`curl -X POST http://localhost:8084/map -H "Content-Type: application/json" -H "X-Request-Id: tid_12345" -H "X-Origin-System-Id: next-video-editor" -d @body.json`

body.json:
```
{
	"_id": "58d8d6cc789d4c000f6b0169",
	"updatedAt": "2017-04-03T16:30:11.106Z",
	"createdAt": "2017-03-27T09:09:32.541Z",
	"mioId": 762380,
	"title": "Trump trade under scrutiny",
	"createdBy": "seb.morton-clark",
	"encoding": {
		"job": 358759376,
		"status": "COMPLETE",
		"outputs": [{
			"audioCodec": "mp3",
			"duration": 65904,
			"mediaType": "audio/mpeg",
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/0x0.mp3"
		},
		{
			"audioCodec": "aac",
			"videoCodec": "h264",
			"duration": 65940,
			"mediaType": "video/mp4",
			"height": 360,
			"width": 640,
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/640x360.mp4"
		},
		{
			"audioCodec": "aac",
			"videoCodec": "h264",
			"duration": 65940,
			"mediaType": "video/mp4",
			"height": 720,
			"width": 1280,
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/1280x720.mp4"
		}]
	},
	"__v": 0,
	"byline": "Filmed by Niclola Stansfield. Produced by Seb Morton-Clark.",
	"description": "Global equities are on the defensive, led by weaker commodities and financials as investors scrutinise the viability of the Trump trade. The FT's Mike Mackenzie reports.",
	"image": "https://api.ft.com/content/ffc60243-2b77-439a-a6c9-0f3603ee5f83",
	"standfirst": "Mike Mackenzie provides analysis of the morning's market news",
	"updatedBy": "seb.morton-clark",
	"isPublished": false,
	"related": [{
		"id": "c4cde316-128c-11e7-80f4-13e067d5072c",
		"title": "Stocks and dollar slide as ‘Trump trade’ fades"
	}],
	"annotations": [{
		"id": "http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740",
		"predicate": "http://www.ft.com/ontology/classification/isClassifiedBy",
	},
	{
		"id": "http://api.ft.com/things/b43f1a91-b805-3453-8c36-1d164c047ca2",
		"predicate": "http://www.ft.com/ontology/annotation/majorMentions",
	},
	{
		"id": "http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325",
		"predicate": "http://www.ft.com/ontology/classification/isClassifiedBy",
	}],
	"encodings": [{
		"mioId": 762380,
		"name": "Trump trade under scrutiny",
		"primary": true,
		"status": "COMPLETE",
		"job": 358759376,
		"outputs": [{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/0x0.mp3",
			"mediaType": "audio/mpeg",
			"duration": 65904,
			"audioCodec": "mp3"
		},
		{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/640x360.mp4",
			"width": 640,
			"height": 360,
			"mediaType": "video/mp4",
			"duration": 65940,
			"videoCodec": "h264",
			"audioCodec": "aac"
		},
		{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/1280x720.mp4",
			"width": 1280,
			"height": 720,
			"mediaType": "video/mp4",
			"duration": 65940,
			"videoCodec": "h264",
			"audioCodec": "aac"
		}]
	}],
	"canBeSyndicated": true,
	"transcription": {
		"status": "COMPLETE",
		"job": "1579674",
		"transcript": "<p>Here's what we're watching with trading underway in London. Global equities under pressure led by weaker commodities and financials as investors scrutinise the viability of the Trump trade. The dollar is weaker. Havens like yen, gold, and government bonds finding buyers. </p><p>As the dust settles over the failure to replace Obamacare, focus now on whether tax reform and other fiscal measures will eventuate. This is where the rubber meets the road for the Trump trade. High flying equity markets had been underpinned by the promise of big tax cuts and fiscal stimulus. And Wall Street is souring. </p><p>One big beneficiary of lower corporate taxes under Trump are small caps. They are now down 2 and 1/2% for the year. While the sector is still much higher since November, this is a key market barometer of prospects for the Trump trade. </p><p>Now while many still think some measure of tax reform or spending will eventuate, markets are very wary, namely of the risk that Congress and the Trump administration fail to reach agreement on legislation, that unlike health care reform, matters a great deal more to investors. </p><p>[MUSIC PLAYING] </p>",
		"captions": [{
			"format": "vtt",
			"url": "https://next-video-editor.ft.com/e2290d14-7e80-4db8-a715-949da4de9a07.vtt",
			"mediaType": "text/vtt"
		}]
	},
	"format": [],
	"type": "video",
	"id":"e2290d14-7e80-4db8-a715-949da4de9a07"
}
```

Return:
200

Body:
```
{
  "uuid": "e2290d14-7e80-4db8-a715-949da4de9a07",
  "annotations": [
    {
      "thing": {
        "id": "http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740",
        "prefLabel": "",
        "predicate": "isClassifiedBy",
        "types": []
      },
      "provenances": [
        {
          "scores": [
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM",
              "value": 0.9
            },
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM",
              "value": 0.9
            }
          ]
        }
      ]
    },
    {
      "thing": {
        "id": "http://api.ft.com/things/b43f1a91-b805-3453-8c36-1d164c047ca2",
        "prefLabel": "",
        "predicate": "majorMentions",
        "types": []
      },
      "provenances": [
        {
          "scores": [
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM",
              "value": 0.9
            },
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM",
              "value": 0.9
            }
          ]
        }
      ]
    },
    {
      "thing": {
        "id": "http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325",
        "prefLabel": "",
        "predicate": "isClassifiedBy",
        "types": []
      },
      "provenances": [
        {
          "scores": [
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-RELEVANCE-SYSTEM",
              "value": 0.9
            },
            {
              "scoringSystem": "http://api.ft.com/scoringsystem/FT-CONFIDENCE-SYSTEM",
              "value": 0.9
            }
          ]
        }
      ]
    },
  ]
}
```

400 - If the mapping couldn't be performed because of invalid provided content.

### Admin endpoints
Healthchecks: [http://localhost:8084/__health](http://localhost:8084/__health)

Ping: [http://localhost:8084/__ping](http://localhost:8084/__ping)

Build-info: [http://localhost:8084/__build-info](http://localhost:8084/__ping)  -  [Documentation on how to generate build-info] (https://github.com/Financial-Times/service-status-go)
