package test

import (
	"io/ioutil"
	"os"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/pkg/errors"
)

func beegoFindConfigFolder() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fs, err := ioutil.ReadDir(wd)
	if err != nil {
		return err
	}

	for _, v := range fs {
		if v.Name() == "conf" {
			return nil
		}
	}

	return os.Chdir("..")
}

func InitBeego() error {
	if err := beegoFindConfigFolder(); err != nil {
		return err
	}

	if err := i18n.SetMessage("uk", "conf/locale_uk-UA.ini"); err != nil {
		if err.Error() != "Lang uk alread exist" {
			return err
		}
	}

	if err := beego.AddFuncMap("i18n", i18n.Tr); err != nil {
		return errors.Wrap(err, "unable to load template i18n tag")
	}

	beego.TestBeegoInit(".")
	beego.BConfig.WebConfig.EnableXSRF = false
	beego.BConfig.WebConfig.Session.SessionOn = false

	return nil
}
