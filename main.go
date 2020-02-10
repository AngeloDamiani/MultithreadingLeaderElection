package main

import (
	"ReactiveSystemProject/LeaderElection"
	"fmt"
	"math/rand"
)


func ringConnection(n int, nodeArray []*LeaderElection.NetworkNode){
	for i:=0;i<n;i++{
		LeaderElection.Connect(nodeArray[i],nodeArray[(i+1)%n])
	}
}


func randomConnection(n int, nodeArray []*LeaderElection.NetworkNode){
	for i:=0;i<n;i++{
		for j:=i+1;j<n;j++{
			if rand.Intn(5)<3{
				LeaderElection.Connect(nodeArray[i],nodeArray[j])
			}
		}
	}
}

func createNetwork(n int)([]*LeaderElection.NetworkNode,*chan int, *chan LeaderElection.Message){

	const masterID  = 0
	const starterID = LeaderElection.StarterID

	var nodeArray []*LeaderElection.NetworkNode

	seed := int64(42)
	rand.Seed(seed)
	p := rand.Perm(n*n*2)
	var IDarr []LeaderElection.Identifier
	exc := 0
	for i:=0;i<n;{
		if p[i+exc] != masterID && p[i+exc] != starterID{
			IDarr = append(IDarr, LeaderElection.Identifier(p[i+exc]))
		} else {
			exc++
			continue
		}
		i++
	}
	IDarr[rand.Intn(len(IDarr))] = starterID

	mutex := make(chan int, 1)
	masterBuff := make(chan LeaderElection.Message, n)

	for i:=0;i<n;i++{
		var node LeaderElection.NetworkNode
		buff := make(chan LeaderElection.Message, n)
		id := IDarr[i]
		node.Init(id, &buff, &mutex, &masterBuff)
		nodeArray = append(nodeArray, &node)
	}
	mutex <- 1

	//ringConnection(n, nodeArray)
	randomConnection(n, nodeArray)

	return nodeArray, &mutex, &masterBuff
}

func printNetwork(nodesArray []*LeaderElection.NetworkNode){
	fmt.Println("-------------------------TOPOLOGY-------------------------")
	for _, v := range nodesArray{
		fmt.Print("[",v.ID)
		fmt.Print("]: My neighbors are ")
		for id, _ := range v.Neighbors{
			fmt.Print("[")
			fmt.Print(id)
			fmt.Print("], ")
		}
		fmt.Print("\b\b \n")
	}
	fmt.Println("----------------------------------------------------------")
}

func main() {


	const nodeNumber = 7


	nodesArray, mutex, masterBuff := createNetwork(nodeNumber)
	printNetwork(nodesArray)

	networkMap := make(map[LeaderElection.Identifier]LeaderElection.Identifier)

	for _, v := range nodesArray{
		networkMap[v.ID] = v.ID
	}


	for _, v := range nodesArray{
		go LeaderElection.RunNode(v)
	}

	for {
		count := 1
		leaderNotification := <-*masterBuff
		networkMap[leaderNotification.Sender] = leaderNotification.Leader
		for k,v := range networkMap {
			if k!=leaderNotification.Sender && v==leaderNotification.Leader{
				count++
			}
		}
		if count == nodeNumber{
			for _, v:= range nodesArray {
				*v.MailBox<-LeaderElection.Message{0,0}
			}
			<- *mutex
			fmt.Println("----------------------------------------------------------")
			fmt.Println("[MASTER]: Convergence reached, Leader is: ", leaderNotification.Leader)
			*mutex <- 1
			break
		}
	}
}
