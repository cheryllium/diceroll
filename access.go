package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "errors"
  "slices"
)

func loadAllowedServers() ([]string, error) {
  type Data struct {
    AllowedServers []string `json:"allowedServers"`
  }

  fileData, err := ioutil.ReadFile("allowed_servers.json")
  if err != nil {
    return []string{}, errors.New(fmt.Sprintf("Error reading file: %s", err))
  }

  var data Data
  if err := json.Unmarshal(fileData, &data); err != nil {
    return []string{}, errors.New(fmt.Sprintf("Error unmarshaling data: %s", err))
  }

  return data.AllowedServers, nil
}

func ServerHasAccess(guildID string) (bool, error) {
  allowed, err := loadAllowedServers()
  if err != nil {
    return false, err
  }

  return slices.Contains(allowed, guildID), nil
}
