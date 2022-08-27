[![Go Report](https://goreportcard.com/badge/github.com/masonkmeyer/nationalize)](https://goreportcard.com/badge/github.com/masonkmeyer/nationalize)
![Build](https://github.com/masonkmeyer/nationalize/actions/workflows/build.yml/badge.svg)

# Nationalize

Nationalize is a go client for the [nationalize.io](https://nationalize.io/) API.


 ## Examples

 You can use this library to call the API client. 
 
 ```golang
client := nationalize.NewClient()
prediction, rateLimit, err := client.Predict("michael")
 ```

This client also supports batch predictions.
