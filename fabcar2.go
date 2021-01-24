/*
TODO LIST:

		April 22, 2020

		1. Client will provide subset of the BDAL JSON to update model and create model
		2. UUID interface will be provided by the javascript. GenerateUUID.js -- UPLOAD JS, Main to deploy type 4
		3. General interface from plugin -- starting connector
		4. To retrieve assets: Give input (Provide the asset id, provide the file name and destination where to put the asset) and get the output (success/fail--Indicator)
		5. Change the approval status for the BDAL model 
		6. JS Interface to IPFS
		7. query a model 
		8. Create ~ as the seperator create a definition 

        May 22,2020

        1. Modify component should take in the entire JSON field and update the fields required. -- updateComponent
        2. Modify Model should take in the entire JSON field and update the fields required. -- updateModel

	
	
		
		
*/

package main

/**
  * Created from archive/fabcar17June2020.go which is the current implemenetation
  */

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

/*
   Model represents a MBSE model.
   A model is stored as a Model struct asset in the blockchain.
   The current version does not implement the model-submodel relationship
   (as shown in the class diagram.)
*/

type Model struct {
	// An universal identifier for a model: e.g. 36ff46e9-23ce-4e1d-96c4-d2f91f802c2b
	ModelId string `json:"ModelId"`
    // Magicdraw identifier 
    SeeMbseId 			string 				`json:"SeeMbseId"`
	// Name should be unique.
	ModelName string `json:"ModelName"`
	// An universal identifier for an organization
	OrgId string `json:"OrgId"`
	// An universal identifier for a project.
	ProjId string `json:"ProjId"`
	// The collection of prescribed components that defines the model.
	// This implements the one to many relationship between 'model'
	// and prescribed components in the class diagram.
	PrescribedComponentList []PrescribedComponent `json:"PrescribedComponentList"`
}

/*
   PrescribedComponent represents an abstract component that is
   prescribed in the MBSE model.
   PrescribedComponent is not stored directly as an asset of the blockchain,
   but as a part of the Model struct.
*/

type PrescribedComponent struct {
	/*
			 An universal identifier for the prescribed component
		     Need to research the proper data type of an universal identifier, such as
		     PCId, ModelId, etc.
		     Not storing child components so no redundant information.
	*/

	// An universal identifier.
	PCId string `json:"PCId"`
	// Name needs not be unique across models.
	PCName string `json:"PCName"`
	CompId string `json:"CompId"`
	// Array of pointers to the prescribed component list
	ChildrenPCList []PrescribedComponent `json:"ChildrenPCList"`
}

/*
   The Component struct represents a MBSE component, such as a SysML diagram.
   Structure tags are used by encoding/json library
*/

type Component struct {

	// A universal identifier for this component.
	CompId string `json:"CompId"`
	// Name provided to the component
	ComponentName string `json:"ComponentName"`
	// Non-nullable foreign key to a model.
	ModelId string `json:"ModelId"`
    // Magicdraw identifier
    SeeMbseId string `json:"SeeMbseId"`

	/*
	   Component relationship.
	*/

	//  The component key for the parent (containing) component.
	//  Simple parent information and simple relationships with
	//  other components.
	ParentCompId        string        `json:"ParentCompId"`
	RelatedToComponents []ToComponent `json:"RelatedToComponents"`

	/*
		Component storage
	*/

	Storage AssetStorage `json:"AssetStorage"`
	//  Author or the component. E.g. 'Bun Yue'
	//  May study Authorship issues for privacy concern.
	//  The following may just be a future direction for study. Owner Organization A
	//  may need to provide the component content to collaboating organization B
	//  but may not want B (a competitor in other areas) to know who
	//  the author is (worrying about poaching, trade secret, etc.)
	//  A possible direction for the future is not to store the personal information
	//  as the author in the blockchain, but instead store a position/role (e.g.
	//  'System Engineer #1') in the blockchain. Organization A can have an
	//  internal mapping of the position to the person. There are pros and cons
	//  and more studies are warranted, but may be in the future.
	Author string `json:"Author"`
	//  Release time for the owner organization on the component.
	//  Default null.
	//  Other options:
	//  (1) Remove entirely. Just use the system time in the blockchain.
	//  (2) Default null, which means using the system time of the block.
	//      One complication, blockchain's system time is for the block, not
	//      individual time for a transaction in the block.
	ReleaseTime string `json:"ReleaseTime"`
	/*
	   Versioning and component status.
	*/
	//  Version of the component. E.g. 2.17.8
	//  {ProjName, OrgName, Title} identifies an component in the MBSE model.
	//  It is a candidate key in the Fabric's world state database, which
	//  keeps the latest state of the component.
	//
	//  In the current design, version can only be increased, and
	//  the state in Fabric world state database always store the
	//  highest version. The BDA module needs to ensure this.
	//  A version represents a baseline in the MBSE project
	//  and every new version needs a fresh approval.
	//  A new subversion, on the other hand, may not need to be
	//  approved. A number of subversions can be put into the
	//  blockchain within an organization for development, or
	//  in a private channel between the developing organization
	//  and (Chief System Integration Engineer (CSIE) for efficient communications.
	Version string `json:"Version"`
	//  Subversion of the component; an integer. E.g. 1.5
	Subversion string `json:"Subversion"`
	//  ComponentStatus: mandatory; default: 'in_model'
	//  Note that Blockchain is immutable and components are not actually deleted.
	//  ComponentStatus can be used to show whether the component has been 'deleted', etc.
	//  Possible values:
	//  [1] 'in_model': the current (latest) version of the component is a
	//      a component of the MBSE model.
	//  [2] 'preliminary': the current version of the component is for
	//      preliminary development and not yet a component in the MBSE model.
	//  [3] 'deprecated': the current version of the component has been replaced
	//      or removed as a component of the MBSE model. An earlier version
	//      is a component of the MBSE model.
	//  [4] 'deleted': the current version of the component is 'deleted'. Since
	//      earlier versions have never been a part of the MBSE model. It
	//      will not be used to construct a digital twin.
	//
	//  A separate, more sophisticated version control model may be needed
	//  in future prototypes.
	ComponentStatus ComponentStatusInfo `json:"ComponentStatus"`
	/*
	   Approval and Review.
	*/
	Approval ApprovalInfo `json:"Approval"`
}

/*
   ToComponent represents a relationship from a component to
   another component.

   This is for experiment. It may be storing this kind of relationship
   in the blockchain may not be effective.

   Implement the 'relates to' association class between two components.
*/
type ToComponent struct {
	// the destination component of the relationship.
	// This is a forward reference as Component has not yet been defined.
	// This is not two structs refer to each other in their definition but
	// it seems GO allow this mutual references of struct.
	ToCompId string `json:"ToCompId"`
	// Relationship Type: e.g. uses, implemented by, ...
	ToCompType string `json:"ToCompType"`
	// Detailed description of the relationship.
	Description string `json:"Description"`
}

/*
   Storage of assets for allowing storage in IPFS.
*/
type AssetStorage struct {
	/*
	   For storage purpose. We keep the term 'Asset' here as storage may
	   apply to assets that are not components.

	   In the current version, AssetStorage applies only to component. Other
	   assets, such as Model, are stored entirely within the blockchain.
	*/

	//  Method of storage of the asset. E.g. 'IPFS', 'direct'
	//  to expand to StorageMethods to allow multiple storage mechanisms.
	//  Current values:
	//     'IPFS': default, to be stored in IPFS.
	//     'direct': to be stored directly in Fabric
	StorageMethod string `json:"StorageMethod"`
	// For components to be stored in a private IPFS.
	// Thus, only meaningful when StorageMethod is 'IPFS'
	// Need: research how to identify the private IPFS and whether and how
	// an identifier (IPFSName, for the moment) will be created and used.
	// Name of the IPFS System to be used. E.g. 'uhcl_IPFS', 'Tietronix_IPFS'
	IPFSName string `json:"IPFSName"`
	// IPFS name.
	//  Multihash address generated by IPFS that can be used to uniquely identify
	//  a file. E.g. QmWWQSuPMS6aXCbZKpEjPHPUZN2NjB3YrhJTHsV4X3vb2t
	AssetIPFSAddress string `json:"AssetIPFSAddress"`
	// IPFS multihash address
	// IPFS files are accessible to anyone with its multihash address.
	// If needed, the files should be encrypted to support the necessary
	// security model.
	// Whether the IPFS component file is encrypted. E.g. TRUE (may be the default value)
	ISEncrypted bool `json:"IsEncrypted"`
	//   Method used for encryption. Only meaningful when IsEnrypted is TRUE.
	//   Default to 'AES256': AES with 256 bits.
	IPFSFileEncMethod string `json:"IPFSFileEncMethod"`
	//   Encryption key to decrypt the encryted file in IPFS. E.g., for AES:
	//   32 bytes: 'q4t7w!z%C*F-JaNdRgUkXp2r5u8x/A?D'
	//   This key is generated for the use of only one file.
	IPFSFileEncKey string `json:"IPFSFileEncKey"`
	//  Asset to be stored within the blockchain directly.
	//  Only meaningful if StorageMethod is 'direct'
	AssetRaw string `json:"AssetRaw"`
	//  component in binary format.
	//  Information about the source files in NBFS (Name-Based File Systems.
	//  If fully provided, this should be
	//  adequate to actually access the source file in the storage system.
	//  Need to further refine SourceType values. E.g. mdzip
	//  Need to develop use cases for supporting the inclusion of
	//  this kind of source information.
	SourceType string `json:"SourceType"`
	//  Name to identify the NBFS. E.g. 'uhcl_localfs_1', 'tietronix_gateway_fs'
	//  Optional as the component may have an unknown provenance.
	//  Need to construct mechanism for mapping SourceName to a unique source.
	SourceName string `json:"SourceName"`
	//  Full file name, including the path.
	//  E.g. gateway/proto1/power/comp1/part32/subpart113.mdzip
	//  Even from an unknown origin, having the filname helps the user
	//  to know what filename to save locally.
	SourceFileName string `json:"SourceFileName"`
	//  Effective start time.
	StartTime string `json:"StartTime"` //time.Time
}

/*
   Status of a component and effective time from when
*/
type ComponentStatusInfo struct {

	//  ComponentStatus: mandatory; default: 'in_model'
	//  Note that Blockchain is immutable and components are not actually deleted.
	//  ComponentStatus can be used to show whether the component has been 'deleted', etc.
	//  Possible values:
	//  [1] 'in_model': the current (latest) version of the component is a
	//      a component of the MBSE model.
	//  [2] 'preliminary': the current version of the component is for
	//      preliminary development and not yet a component in the MBSE model.
	//  [3] 'deprecated': the current version of the component has been replaced
	//      or removed as a component of the MBSE model. An earlier version
	//      is a component of the MBSE model.
	//  [4] 'deleted': the current version of the component is 'deleted'. Since
	//      earlier versions have never been a part of the MBSE model. It
	//      will not be used to construct a digital twin.
	//  [5] 'final': no more change can be made.
	//
	//  A separate, more sophisticated version control model may be needed
	//  in future prototypes.
	ComponentStatus string `json:"ComponentStatus"`
	StatusSince     string `json:"StatusSince"` //time.Time
}

/*
   Approval and Review.
*/
type ApprovalInfo struct {
	//  Approval information. Optional.
	//  E.g. 'disapproved': no, 'approved as written', 'approved with modification',
	//  'rework and resubmit'
	//  default value of 'started'.
	ApprovalStatus string `json:"ApprovalStatus"`
	//  Time of approval. E.g. 11:22:15.2021 (or Unix time)
	//  need to research Go and Json time data type.
	ApprovalTime string `json:"ApprovalTime"` //time.Time
	//  The user who approves.
	//  Need to research how to link to the user defined by the membership
	//  service of Fabric.
	Approver string `json:"Approver"`
	//  An approval status update is requested: the requestor.
	//  Need to research how to link to the user defined by the member
	//  service of Fabric
	StatusUpdateRequestor string `json:"StatusUpdateRequestor"`
	//  Request to update to which status.
	StatusUpdateTo string `json:"StatusUpdateTo"`
	// Time of the update request.
	StatusUpdateReqTime string `json:"StatusUpdateReqTime"` //time.Time
}

type SmartContract struct {
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryModel" {
		return s.queryModel(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createModel" {
		return s.createModel(APIstub, args)
	} else if function == "createComponent" {
		return s.createComponent(APIstub, args)
	} else if function == "queryModelbyId" {
		return s.queryModelbyId(APIstub, args)
	} else if function == "queryComponentbyId" {
		return s.queryComponentbyId(APIstub, args)
	} else if function == "richQueryModel" {
		return s.richQueryModel(APIstub, args)
	} else if function == "modifyComponent" {
		return s.modifyComponent(APIstub, args)
	} else if function == "queryAllModels" {
		return s.queryAllModels(APIstub)
	} else if function == "queryAllComponents" {
		return s.queryAllComponents(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryModel(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ModelAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(ModelAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {

	return shim.Success(nil)
}

func (s *SmartContract) createModel(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	indexName := "model~id"
	var result Model
	if err := json.Unmarshal([]byte(args[0]), &result); err != nil {
		return shim.Error(err.Error())
	}

	ModelAsBytes, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}

	attributeIdIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{"model", result.ModelId})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(attributeIdIndexKey, ModelAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (s *SmartContract) createComponent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Expecting number of arguments: 1")
	}
	indexName := "component~id"
	var result Component
	if err := json.Unmarshal([]byte(args[0]), &result); err != nil {
		return shim.Error(err.Error())
	}
	attributeIdIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{"component", result.CompId})
	if err != nil {
		return shim.Error(err.Error())
	}

	ComponentAsBytes, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(attributeIdIndexKey, ComponentAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryModelbyId(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	indexName := "model~id"

	attrIdIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{"model", args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	ModelAsBytes, err := APIstub.GetState(attrIdIndexKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ModelAsBytes)
}

func (s *SmartContract) queryComponentbyId(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	indexName := "component~id"

	attrIdIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{"component", args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	ComponentAsBytes, err := APIstub.GetState(attrIdIndexKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ComponentAsBytes)
}

func (s *SmartContract) richQueryModel(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key and value to query")
	}

	queryString := string(args[0])
	//fmt.Printf(queryString)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for }"
		return shim.Error(jsonResp)
	} else if string(queryResults) == "[]" {
		jsonResp := []byte("{\"Error\":\"Model does not exist: ")
		return shim.Success(jsonResp)
	}

	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	//fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString) //USE ONLY FOR DEBUGGING

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String()) // USE ONLY FOR DEBUGGING

	return buffer.Bytes(), nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

//Functions for each of the components
// Functions of the components
func ComponentNameChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.ComponentName = incomingData.ComponentName

	return bcJsonData
}

/* Can be added later if needed

func ModelIdChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.ModelId = incomingData.ModelId

	return bcJsonData
}

func ParentCompIdChange(incomingData Component, bcJsonData Component) Component {
	bcJsonData.ParentCompId = incomingData.ParentCompId

	return bcJsonData
}
*/
func RelatedToComponentsChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.RelatedToComponents = append(incomingData.RelatedToComponents)

	return bcJsonData
}

func AssetStorageChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.Storage.StorageMethod = incomingData.Storage.StorageMethod
	bcJsonData.Storage.IPFSName = incomingData.Storage.IPFSName
	bcJsonData.Storage.AssetIPFSAddress = incomingData.Storage.AssetIPFSAddress
	bcJsonData.Storage.ISEncrypted = incomingData.Storage.ISEncrypted
	bcJsonData.Storage.IPFSFileEncMethod = incomingData.Storage.IPFSFileEncMethod
	bcJsonData.Storage.IPFSFileEncKey = incomingData.Storage.IPFSFileEncKey
	bcJsonData.Storage.AssetRaw = incomingData.Storage.AssetRaw
	bcJsonData.Storage.SourceType = incomingData.Storage.SourceType
	bcJsonData.Storage.SourceName = incomingData.Storage.SourceName
	bcJsonData.Storage.SourceFileName = incomingData.Storage.SourceFileName
	bcJsonData.Storage.StartTime = incomingData.Storage.StartTime

	return bcJsonData
}

func AuthorChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.Author = incomingData.Author

	return bcJsonData
}

func ReleaseTimeChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.ReleaseTime = incomingData.ReleaseTime

	return bcJsonData
}

func VersionChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.Version = incomingData.Version

	return bcJsonData
}

func SubversionChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.Subversion = incomingData.Subversion

	return bcJsonData
}

func ComponentStatusInfoChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.ComponentStatus.ComponentStatus = incomingData.ComponentStatus.ComponentStatus
	bcJsonData.ComponentStatus.StatusSince = incomingData.ComponentStatus.StatusSince

	return bcJsonData
}

func approvalChange(incomingData Component, bcJsonData Component) Component {

	bcJsonData.Approval.ApprovalStatus = incomingData.Approval.ApprovalStatus
	bcJsonData.Approval.ApprovalTime = incomingData.Approval.ApprovalTime
	bcJsonData.Approval.Approver = incomingData.Approval.Approver
	bcJsonData.Approval.StatusUpdateRequestor = incomingData.Approval.StatusUpdateRequestor
	bcJsonData.Approval.StatusUpdateTo = incomingData.Approval.StatusUpdateTo
	bcJsonData.Approval.StatusUpdateReqTime = incomingData.Approval.StatusUpdateReqTime

	return bcJsonData
}

// ModifyComponentField --- merged with ModifyComponent should get the entire component JSON

func (s *SmartContract) modifyComponent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	indexName := "component~id"

	var incomingJSON Component

	if err := json.Unmarshal([]byte(args[0]), &incomingJSON); err != nil {
		return shim.Error(err.Error())
	}

	attrIdIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{"component", incomingJSON.CompId})
	if err != nil {
		return shim.Error(err.Error())
	}

	ComponentAsBytes, _ := APIstub.GetState(attrIdIndexKey)

	var changedJSON Component
	if err := json.Unmarshal(ComponentAsBytes, &changedJSON); err != nil {
		return shim.Error(err.Error())
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	//get the top level keys of the json
	incomingDataMap := make(map[string]Component)
	json.Unmarshal([]byte(args[0]), &incomingDataMap)

	for key := range incomingDataMap {
		//fmt.Println(key)
		if key == "ComponentName" {
			changedJSON = ComponentNameChange(incomingJSON, changedJSON)
		} else if key == "RelatedToComponents" {
			changedJSON = RelatedToComponentsChange(incomingJSON, changedJSON)
		} else if key == "AssetStorage" {
			changedJSON = AssetStorageChange(incomingJSON, changedJSON)
		} else if key == "Author" {
			changedJSON = AuthorChange(incomingJSON, changedJSON)
		} else if key == "ReleaseTime" {
			changedJSON = ReleaseTimeChange(incomingJSON, changedJSON)
		} else if key == "Version" {
			changedJSON = VersionChange(incomingJSON, changedJSON)
		} else if key == "Subversion" {
			changedJSON = SubversionChange(incomingJSON, changedJSON)
		} else if key == "ComponentStatus" {
			changedJSON = ComponentStatusInfoChange(incomingJSON, changedJSON)
		} else if key == "Approval" {
			changedJSON = approvalChange(incomingJSON, changedJSON)
		} /* else if key == "ModelId" {
			changedJSON = ModelIdChange(incomingJSON, changedJSON)
		} else if key == "ParentCompId" {
			changedJSON = ParentCompIdChange(incomingJSON, changedJSON)
		} */ /* TODO: what about SeeMbseId value change? */
	}

	finalChangedJSON, errMar := json.Marshal(changedJSON)
	if errMar != nil {
		return shim.Error(errMar.Error())
	}

	if err := APIstub.PutState(attrIdIndexKey, finalChangedJSON); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

func (s *SmartContract) queryAllModels(stub shim.ChaincodeStubInterface) sc.Response {
	indexName := "model~id"
	modelResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer modelResultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(modelResultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	//fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String()) // USE ONLY FOR DEBUGGING

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryAllComponents(stub shim.ChaincodeStubInterface) sc.Response {
	indexName := "component~id"
	componentResultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer componentResultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(componentResultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	//fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String()) // USE ONLY FOR DEBUGGING

	return shim.Success(buffer.Bytes())
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
