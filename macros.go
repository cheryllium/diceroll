package main

import (
  "gorm.io/gorm"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm/logger"
  "log"
  "errors"
)

var db *gorm.DB = nil

type Macro struct {
  gorm.Model
  Guild string
  Name string
  Expression string
}

func InitDB() {
  log.SetPrefix("macros: ")
  log.SetFlags(0)
  
  var err error
  db, err = gorm.Open(sqlite.Open("macros.db"), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Silent),
  })
  if err != nil {
    log.Fatal(err)
  }

  db.AutoMigrate(&Macro{})
}

func FindMacro(guild string, name string) (*Macro, error) {
  var macro Macro

  result := db.Where("Guild = ? AND Name = ?", guild, name).First(&macro)
  if result.Error != nil {
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("No rows found")
		}
		return nil, errors.New("Database error")
  }

  return &macro, nil
}

func MakeMacro(macro *Macro) {
  db.Create(macro)
}

func DeleteMacro(macro *Macro) {
  db.Delete(macro)
}

func EditMacro(macro *Macro) {
  db.Save(&macro)
}

func ListMacros(guild string) ([]Macro, error) {
  var macros []Macro
  result := db.Where("Guild = ?", guild).Order("Name").Find(&macros)
  if result.Error != nil {
    return nil, errors.New("Database error (possibly no rows found)")
  }
  return macros, nil
}

func ListMacrosByName(name string) ([]Macro, error) {
  var macros []Macro
  result := db.Where("Name = ?", name).Find(&macros)
  if result.Error != nil {
    return nil, errors.New("Database error (possibly no rows found)")
  }
  return macros, nil
}
