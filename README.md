# GoConcurrency-Bootcamp-2022
 
## Introduction
 
Thank you for participating in the Advanced GO Bootcamp course!
Here, you'll find instructions for completing your certification.
 
## The Challenge
 
The purpose of the challenge is for you to demonstrate your GO skills. This is your chance to show off everything you've learned during the course!!
 
You will build upon an existing Go project that has many problems that concurrency can solve. You will solve those problems using two of the following concurrency patterns:
- Generator
- Fan in - Fan out
- Pipeline
 
## Requirements
 
These are the main requirements we will evaluate:
 
- Implement two of the three concurrency patterns proposed above (the three is a nice to have)
- In case of any error, stop the whole concurrency process
- Improve significantly the process time of the endpoints
 
## Getting Started
 
To get started, follow these steps:
 
1. Fork this project
2. Commit periodically
3. Apply changes according to the reviewer's comments
4. Have fun!
 
> Important: You need to have installed on your box Go v1.16 or above and redis
 
## Deliverables

You can submit the whole project at once and request feedback or you can split the challenge applying the selected patterns for every stage and check with your mentor. It's up to you.
 
## The current state of the application and the problems
 
The application is an API that interacts with 'https://pokeapi.co/api/v2/'. This app has three main functions which are exposed in the following endpoints:
 
1. (POST) /api/provide
 - Fills the local CSV with data from PokeAPI. It fetches the information for the pokemons with ID 1 to 10 by default.
 - You can specify the range of IDs of the pokemons to fetch like this:
```
{
 "from": 10,
 "to": 50
}
```
 - The problem with this process is every id takes ~2.5 seconds in process, so if we try to process 50 ids, it will take more than 2 minutes in process and that time increase proportionally to the number of ids
2. (PUT) /api/refresh-cache
 - Recovers the pokemon information from the CSV and for each pokemon it fetches and feeds it to the struct.
 - Saves the complete pokemon information in redis.
 - This endpoint has two problems:
   1. We need to read the whole csv and then we need to hit the abilities endpoint, so this process can take long time to finish (~19 seconds per 40 records)
   2. This process is not asynchronous, so we need to finish step by step every stage (read, fetch, feed and save)
3. (GET) /api/pokemons
 - Returns all the pokemons in cache (this is only for visualization).
 
## Deliverable (due Friday May 13th, 23:59PM)
 
Based on the self-study material and mentorship covered until this deliverable, we suggest you perform the following:
 
1. Run the project, play with it and understand the problem with the API
2. Select two of the following patterns in order to solve the corresponding problem:
- Generator: This pattern can be used to handle multiple requests with potential pararellism. You can implement it in the /provide endpoint to hit the multiple endpoints asynchronously.
 *(Goal: Process 100 ids in less than 3 seconds)*
 - Fan in - Fan out: These two patterns can be used together for handle multiple inputs (fan-in) and outputs (fan-out). You can implement these patterns in /refresh-cache in order to read the csv line by line and hitting to the abilities endpoint concurrently.
 *(Goal: Process 40 records with 3 workers in less than 6 seconds*)
 - Pipeline - In order to make the whole process asynchronous and write to the cache by batches at the same time that we are reading the csv file, we can implement the pipeline pattern to do the following steps asynchronous:
   Read the csv - Feed the pokemons with abilities - Save into cache. ***(For this pattern is required to process the pokemons by batches instead all at once)***
 *(Goal: Write in cache in real time by batches while the file is reading)*
3. Implement one pattern at a time, generate a pull request and ask for feedback to your mentor while you are developing the next pattern
 
At the end you need to:
- Solve two of the three problems presented on the API with concurrency patterns
- Present a comparison table of the time processes
 
> Important: In case of any error, the whole process needs to stop (hint: channels) except by the pipeline pattern. In that case, we need to stop but it's fine to have some batches processed.
 
## Final Deliverable (due Wednesday June 8th, 2:00PM)
> Important: this is the final deliverable, so all the requirements must be included.
 
## Submitting the deliverables
 
For submitting your work, you should follow these steps:
 
1. Create a pull request with your code, targeting the master branch of your fork.
2. Fill this [form](https://forms.gle/urV6szfnCVMqp4UL9) including the PR’s url
3. Stay tune for feedback
4. Do the changes according to the reviewer's comments
 
## Documentation
 
### Must to learn

### Self-Study Material
 
- [Golang Concurrency Patterns](https://www.karanpratapsingh.com/courses/go/advanced-concurrency-patterns)
- [Pipelines and cancellation](https://go.dev/blog/pipelines)
