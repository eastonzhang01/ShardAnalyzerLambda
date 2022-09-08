package main

import (
	"shardanalyzer/config"
	"shardanalyzer/reports"
	"shardanalyzer/models"
	
	"github.com/aws/aws-lambda-go/lambda"																						
	"github.com/aws/aws-lambda-go/events"
	"encoding/json"
	// "errors"
	
	"net/http"
	"io/ioutil"
	
	// "log"
	log "github.com/sirupsen/logrus"
)
																												
func main(){																													
	lambda.Start(Handler)
}

type InputEvent struct {
	AvailabilityZones int `json:"availabilityzones"`
	ClusterName string `json:"clustername"`
	ClientName string `json:"clientname"`
	Search bool `json:"search"`
	TargetSize int `json:"targetsize"`
	DomainEndpoint string `json:"domainendpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
	CatShards string `json:"rawInput"`
}

type ResponseJson struct {
	Title                            string                       			`json:"title"`
	ClusterName                      string                       			`json:"cluster_name"`
	NumberOfAZs                      int                          			`json:"number_of_azs"`
	NumberOfDataNodes                int                          			`json:"number_of_data_nodes"`
	TotalPrimarySize                 int64                        			`json:"total_primary_size"`
	TotalReplicaSize                 int64                        			`json:"total_replica_size"`
	TotalShards                      int                          			`json:"total_shards"`
	PotentialShards                  int                          			`json:"potential_shards"`
	TotalIndices					 int						  			`json:"total_indices"`									// Index Count is from summing length of each Indices array within the IndexPatternRecommendation that is within the IndexPatternRecommendationRollup array
	TotalIndexPatterns				 int						  			`json:"total_index_patterns"`							// number of IndexPatternRecommendation structs that have Pattern!=No Patterns
	RecommendedShardSizeInGb         int                          			`json:"recommended_shard_size_in_gb"`
	NeedAdjustment         			 bool                          			`json:"need_adjustment"`
	NodeStats         			 	 []models.NodeStats           			`json:"node_stats"`
	LargeIndices			         []models.IndexRecommendation           `json:"large_indices"`									// Array of Indices with shards over 50g
	IndexPatternRecommendationRollup []models.IndexPatternRecommendation 	`json:"index_pattern_recommendation_rollup"`			// Array of IndexPatternRecommendation structs
	EmptyIndices                     []string                     			`json:"empty_indices,omitempty"`						// Array of strings listing the empty indices	
}

type logResponse struct {
	ClusterName                      string                       			`json:"cluster_name"`
	NumberOfAZs                      int                          			`json:"number_of_azs"`
	NumberOfDataNodes                int                          			`json:"number_of_data_nodes"`
	TotalPrimarySize                 int64                        			`json:"total_primary_size"`
	TotalReplicaSize                 int64                        			`json:"total_replica_size"`
	TotalShards                      int                          			`json:"total_shards"`
	PotentialShards                  int                          			`json:"potential_shards"`
	TotalIndices					 int						  			`json:"total_indices"`
	TotalIndexPatterns				 int						  			`json:"total_index_patterns"`
	RecommendedShardSizeInGb         int                          			`json:"recommended_shard_size_in_gb"`
	NeedAdjustment         			 bool                          			`json:"need_adjustment"`
}

type infoLog struct {
	Status						string				`json:"status"`
	InputCatShards				bool				`json:"input_cat/shards"`
	DomainEndpoint				string				`json:"domainendpoint"`
	AvailabilityZones			int					`json:"availabilityzones"`
	ClusterName 				string 				`json:"clustername"`
	ClientName 					string 				`json:"clientname"`
	Search 						bool 				`json:"search"`
	TargetSize 					int 				`json:"targetsize"`
	ErrorMessage				string				`json:"errorMessage,omitempty"`
	NumberOfDataNodes			int                 `json:"number_of_data_nodes,omitempty"`
	TotalPrimarySize            int64               `json:"total_primary_size,omitempty"`
	TotalReplicaSize            int64          		`json:"total_replica_size,omitempty"`
	TotalShards                 int            		`json:"total_shards,omitempty"`
	PotentialShards             int            		`json:"potential_shards,omitempty"`
	TotalIndices				int					`json:"total_indices,omitempty"`
	TotalIndexPatterns			int					`json:"total_index_patterns,omitempty"`
	NeedAdjustment         		bool                `json:"need_adjustment,omitempty"`
}

const ERROR = "ERROR"
const SUCCESS = "SUCCESS"

// create just one log structure, add omit empty to all JSON fields 	
// create string field is error, response, request 
// just have one log struct and only send fields that are not empty
// log.debug message 
// log.trace
// log.info and log.error
// one log object with type, 
// log.debug and log.trace send lot of information, use to understand workflow
// for basic stats, just need input, output and status
// generate error log in multiple places
// status, response object, request object, error object 
// when success fill all except error object, when have failure fill all except response object 

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.SetFormatter(&log.JSONFormatter{})
	
	HEAD := map[string]string{
		"Access-Control-Allow-Origin":  "*",
        "Access-Control-Allow-Methods": "DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT",
        "Access-Control-Allow-Headers": "X-Amz-Date,X-Api-Key,X-Amz-Security-Token,X-Requested-With,X-Auth-Token,Referer,User-Agent,Origin,Content-Type,Authorization,Accept,Access-Control-Allow-Methods,Access-Control-Allow-Origin,Access-Control-Allow-Headers",	
    }

	if request.HTTPMethod == "POST" {
		event, handleRequestBodyError := handleRequestBody(request.Body)
		if handleRequestBodyError != "" {
			createLogError(handleRequestBodyError, event)
			return events.APIGatewayProxyResponse{						// return Error respnse
				Headers: 		HEAD,
				Body:			handleRequestBodyError,
				StatusCode:		400}, nil
		}
		
		catShardsOutput, shardInputError := getCatShards(event)
		if shardInputError != "" {
			createLogError(shardInputError, event)
			return events.APIGatewayProxyResponse{						// return Error respnse
				Headers: 		HEAD,
				Body:			shardInputError,
				StatusCode:		400}, nil
		}
		
		args := config.ShardRecommendationRequest{						// Create struct of all input info
			CatShards:         catShardsOutput,
			TargetShardSizeGB: event.TargetSize,
			NumberOfAzs:       event.AvailabilityZones,
			IsSearchWorkload:  event.Search,
			ClusterName:       event.ClientName,
			ClientName:        event.ClusterName,						// For some Reason cluster name and client name need to be switched?
		}
		
		cluster, err := args.ParseStats()								// parses cat/shards input and validates
		if err != nil {
			parseError := "ERROR: error occured in parsing stats"
			createLogError(parseError, event)
			return events.APIGatewayProxyResponse{						// return events.APIGatewayProxyResponse
				Headers: 		HEAD,
				Body:			parseError,
				StatusCode:		400}, nil
		}
		println("total shards", len(cluster.Rollup))
		recommendation := cluster.PrepareRecommendation()				// recommendation is JSON struct that is outputted
		reports.MergeSingleIndexPatterns(&recommendation)
		
		var nodeArray []models.NodeStats
		for _, ns := range cluster.Nodes {								// copy nodes from map into an array
			nodeArray = append(nodeArray, *ns)
		}
		
		_, indicesGreaterFifty := recommendation.GetIndicesWithLargerShards(50)
		
		finalResponse := ResponseJson{
			Title: 								recommendation.Title,
			ClusterName:						recommendation.ClusterName,
			NumberOfAZs:						recommendation.NumberOfAZs,
			NumberOfDataNodes:					recommendation.NumberOfDataNodes,
			TotalPrimarySize:					recommendation.TotalPrimarySize,
			TotalReplicaSize:					recommendation.TotalReplicaSize,
			TotalShards:						recommendation.TotalShards,
			PotentialShards:					recommendation.PotentialShards,
			TotalIndices:						recommendation.GetIndexCount(),
			TotalIndexPatterns:					recommendation.GetTotalIndexPatterns(),
			RecommendedShardSizeInGb:			recommendation.RecommendedShardSizeInGb,
			LargeIndices:						indicesGreaterFifty,
			NeedAdjustment:						recommendation.NeedsShardAdjustment(),
			NodeStats:							nodeArray,
			IndexPatternRecommendationRollup:	recommendation.IndexPatternRecommendationRollup,
			EmptyIndices:						recommendation.EmptyIndices,
		}
		
		
		bodyBytes, err := json.Marshal(finalResponse)
		if err != nil {
			marshalError := "ERROR: error occured in json.Marshal of the final response"
			createLogError(marshalError, event)
			return events.APIGatewayProxyResponse{						// return events.APIGatewayProxyResponse
				Headers: 		HEAD,
				Body:			marshalError,
				StatusCode:		400}, nil
		}
		
		createLogResponse(event, finalResponse)
		return events.APIGatewayProxyResponse{							// return events.APIGatewayProxyResponse
			Headers: 		HEAD,
			Body:			string(bodyBytes),
			StatusCode:		200}, nil
		
	} else{
		notPostError := "ERROR: made a non-POST request"
		log.WithFields(log.Fields{
			"details": request,
		}).Error(notPostError)
		return events.APIGatewayProxyResponse{							// return events.APIGatewayProxyResponse
			Headers: 		HEAD,
			Body:			notPostError,
			StatusCode:		400}, nil
	}
}

func createLogError(errorMessage string, event InputEvent){
	errorStruct:= infoLog{
				Status: ERROR,
				InputCatShards: event.CatShards!="",
				DomainEndpoint: event.DomainEndpoint,
				AvailabilityZones: event.AvailabilityZones,
				ClusterName: event.ClusterName,
				ClientName: event.ClientName,
				Search: event.Search,
				TargetSize: event.TargetSize,
				ErrorMessage: errorMessage,
			}
	log.WithFields(log.Fields{
		"Details": errorStruct,
	}).Error(errorMessage)
}

func handleRequestBody(requestBody string)(InputEvent, string){
	event:=InputEvent{}
	err := json.Unmarshal([]byte(requestBody), &event) 					// change request body into InputEvent struct
	if err != nil {
		return InputEvent{}, "ERROR: error occured in json.Unmarshal of request Body"
	}
	
	return event, ""
}

func getCatShards(event InputEvent)(string, string) {
	if(event.DomainEndpoint != ""){										// first check if Domain endpoint is not empty, then use GET request for _cat/shards
		getRequestReturn, err := getURL(event.DomainEndpoint, event.Username, event.Password)
		if err != ""{													// check if error occured
			return "", err												// if error occured pass along error message
		} else{
			return getRequestReturn, ""									// if no error occured pass along GET request return
		}
	} else if(event.CatShards != ""){									// if domain endpoint is empty check if cat/shards is empty
		return event.CatShards, ""
	} else {															// if both are empty then return an error message
		return "", "ERROR: Empty _cat/shards input and Domain Endpoint"
	}
}

func getURL(url string, username string, password string) (string, string) {
	client := &http.Client{}
	
	// check if there is / at end
	if(string(url[len(url)-1])=="/"){
		url=url+"_cat/shards?v"
	} else{
		url = url+"/_cat/shards?v"
	}
	// create new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil{
		return "", "ERROR: error occured in creating Request"
		// RETURN ERROR for creating request
	}
	// set authentication of new request
	if(username != "" && password != ""){								// if one of the fields is empty, don't set basic authentication
		req.SetBasicAuth(username, password)
	}
	// send request
	resp, err := client.Do(req)
	if err != nil{
		return "", "ERROR: error occured in sending GET request to the DomainEndpoint"
		// RETURN ERROR for sending request
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
    s := string(bodyText)
	
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
        return s, ""
    } else {
        return "", "ERROR: " + s
    }
	// resp.getStatusCode != HTTP.OKstatus
	// then generate error response
	// whether 200 or error, read the response body 
	// if its an error, send the response body as an error message 
	// not an error then proceed as normal with response body
}

func createLogResponse(event InputEvent, response ResponseJson) {
	
	successStruct:= infoLog{
				Status: SUCCESS,
				InputCatShards: event.CatShards!="",
				DomainEndpoint: event.DomainEndpoint,
				AvailabilityZones: event.AvailabilityZones,
				ClusterName: event.ClusterName,
				ClientName: event.ClientName,
				Search: event.Search,
				TargetSize: event.TargetSize,
				NumberOfDataNodes: response.NumberOfDataNodes,
				TotalPrimarySize: response.TotalPrimarySize,
				TotalReplicaSize: response.TotalReplicaSize,
				TotalShards: response.TotalShards,
				PotentialShards: response.PotentialShards,
				TotalIndices: response.TotalIndices,
				TotalIndexPatterns: response.TotalIndexPatterns,
				NeedAdjustment: response.NeedAdjustment,
			}
	log.WithFields(log.Fields{
		"Details": successStruct,
	}).Info("Successful Execution")
}

