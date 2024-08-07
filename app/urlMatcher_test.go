package main

import (
	"testing"
)

func Test_Match_String(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/")
	if _, ok := matcher("/qwe/"); !ok {
		t.Error("Should be equal")
	}
}

func Test_Match_String_Negative(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/")
	if _, ok := matcher("/kjdfg/"); ok {
		t.Error("Shouldn't be equal")
	}
}

func Test_Match_String_Negative1(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe")
	if _, ok := matcher("/qwer"); ok {
		t.Error("Shouldn't be equal")
	}
}

func Test_Match_StringAndParams(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/{param1}/")
	if params, ok := matcher("/qwe/asd/"); !ok {
		t.Error("Should be equal")
	} else if params["param1"] != "asd" {
		t.Error("Should be equal to asd")
	}
}

func Test_Match_StringAndParams2(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/{param1}/xcv")
	if params, ok := matcher("/qwe/asd/xcv"); !ok {
		t.Error("Should be equal")
	} else if params["param1"] != "asd" {
		t.Error("Should be equal to asd")
	}
}

func Test_Match_StringAndParams3(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/{param1}/{param2}")
	if params, ok := matcher("/qwe/asd/xcv"); !ok {
		t.Error("Should be equal")
	} else if params["param1"] != "asd" {
		t.Error("Should be equal to asd")
	} else if params["param2"] != "xcv" {
		t.Error("Should be equal to xcv")
	}
}

func Test_Match_StringAndParams4(t *testing.T) {
	matcher, _ := CreateUrlMatcher("/qwe/{*param1}")
	if params, ok := matcher("/qwe/asd/xcv"); !ok {
		t.Error("Should be equal")
	} else if params["param1"] != "asd/xcv" {
		t.Error("Should be equal to asd/xcv")
	}
}
