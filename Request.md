# free5gc test Modify

* Add this part into ```registration_test.go```
```golang
//OCF Testing
func CTF(ue_ID string){
	values := map[string]string{"ue_ID": ue_ID}
    json_data, err := json.Marshal(values)

    resp, err := http.Post("https://je752rauad.execute-api.us-east-1.amazonaws.com/Nchf/registration", "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Println("[Registration] API Failed.")
    }

	body, err := ioutil.ReadAll(resp.Body)

   	sb := string(body)
   	log.Printf(sb)

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)
    log.Println(res["json"])

	Nchf_ConvergedChargingFunction_create(sb)
}

//Write session data into UE database
func Nchf_ConvergedChargingFunction_create(ue_ID string){
	values := map[string]string{"ue_ID": ue_ID}
    json_data, err := json.Marshal(values)

    resp, err := http.Post("https://je752rauad.execute-api.us-east-1.amazonaws.com/Nchf/create", "application/json",
        bytes.NewBuffer(json_data))

    if err != nil {
        log.Println("[Create] API Failed.")
    }

	log.Println("GU Authorized.")

	body, err := ioutil.ReadAll(resp.Body)

	sb := string(body)
   	log.Printf(sb)

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res["json"])
	time.Sleep(5 * time.Second)
	Write_Session(sb)
}

//Write session data into UE database
func Write_Session(ue_ID string){
	values := map[string]string{"ue_ID": ue_ID}
    json_data, err := json.Marshal(values)

	resp, err := http.Post("https://je752rauad.execute-api.us-east-1.amazonaws.com/Nchf/continous-write", "application/json",
	bytes.NewBuffer(json_data))
	
	if err != nil {
        log.Println("[Session] API Failed.")
    }

	body, err := ioutil.ReadAll(resp.Body)

	sb := string(body)
   	log.Printf(sb)

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res["json"])

	log.Println("Session Started...")
}

//Update user GU
func Nchf_ConvergedChargingFunction_update(ue_ID string){
	values := map[string]string{"ue_ID": ue_ID}
    json_data, err := json.Marshal(values)

	resp, err := http.Post("https://je752rauad.execute-api.us-east-1.amazonaws.com/Nchf/update", "application/json",
	bytes.NewBuffer(json_data))
	
	if err != nil {
        log.Println("[Update] API Failed.")
    }

	log.Println("GU Updated!!!")

	body, err := ioutil.ReadAll(resp.Body)

	sb := string(body)
   	log.Printf(sb)

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res["json"])
}

//Poll out the session data to S3, and Delete the session data
func Nchf_ConvergedChargingFunction_release(ue_ID string){
	values := map[string]string{"ue_ID": ue_ID}
    json_data, err := json.Marshal(values)

	resp, err := http.Post("https://je752rauad.execute-api.us-east-1.amazonaws.com/Nchf/release", "application/json",
	bytes.NewBuffer(json_data))
	
	if err != nil {
        log.Println("[Release] API Failed.")
    }

	log.Println("Session Released!!!")

	body, err := ioutil.ReadAll(resp.Body)

	sb := string(body)
   	log.Printf(sb)

    var res map[string]interface{}

    json.NewDecoder(resp.Body).Decode(&res)
    fmt.Println(res["json"])
}
```

* Add this part into ```func TestRegistration()```

```golang
//CTF Test
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(10000)
	var ueID string = "ue-" + strconv.Itoa(randNum)
	log.Println("CTF Test Started...")
	CTF(ueID)
	Nchf_ConvergedChargingFunction_release(ueID)
	Nchf_ConvergedChargingFunction_update(ueID)
```