package net

// func SyncChain(ledger *Ledger) error {
// 	for _, node := range ledger.Nodes {
// 		resp, err := http.Get(node.Address + "/chain")
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		bodyBytes, err := ioutil.ReadAll(resp)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		log.Println(string(bodyBytes))
// 	}

// 	return nil
// }
