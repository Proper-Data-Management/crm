package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/go-ldap/ldap"
	"github.com/julienschmidt/httprouter"
)

func LdapAuthByLdapId(ldapId int64, userName string, password string) error {

	o := orm.NewOrm()
	o.Using("default")
	person_dn, host, pre := "", "", ""
	err := o.Raw(DbBindReplace("select person_dn,host,pre from ldaps where id=?"), ldapId).QueryRow(&person_dn, &host, &pre)
	if err != nil {
		return err
	}
	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}

	err = l.Bind("cn="+userName+","+person_dn, password)
	if err != nil {

		log.Println("Failed To Connect Ldap. Trying to connect Active Directory")
		err = l.Bind(pre+userName, password)
		if err != nil {
			return err
		} else {
			return nil
		}

		return err
	}
	return nil
}

func LdapSync(res http.ResponseWriter, req *http.Request, params httprouter.Params) {

	ldapUUID := params.ByName("ldap")

	type tResult struct {
		Result string `json:"result"`
	}
	result := tResult{Result: "ok"}

	err := LdapSyncAd(ldapUUID)

	if err != nil {

		result.Result = err.Error()
	}
	resP, _ := json.Marshal(result)
	fmt.Fprint(res, string(resP))

}

func LdapSyncLdap(res http.ResponseWriter, req *http.Request, params httprouter.Params) {

	ldapId := params.ByName("ldap_id")
	o := orm.NewOrm()
	o.Using("default")
	admin_username, admin_password, person_dn, host := "", "", "", ""
	err := o.Raw(DbBindReplace("select admin_username,admin_password,person_dn,host from ldaps where id=?"), ldapId).QueryRow(&admin_username, &admin_password, &person_dn, &host)
	if err != nil {
		return
	}
	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return
	}
	err = l.Bind(admin_username, admin_password)
	if err != nil {
		return
	}
	attributes := []string{"cn"}
	filter := "(cn=*)"
	search := ldap.NewSearchRequest(
		person_dn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil)
	sr, _ := l.Search(search)
	for _, v := range sr.Entries {
		rs, err := o.Raw(DbBindReplace("insert into users (title,email,ldap_id)values(?,?,?)"), v.GetAttributeValue("cn"), v.GetAttributeValue("cn"), ldapId).Exec()
		if err == nil {
			lId, _ := rs.LastInsertId()
			o.Raw(DbBindReplace("insert into user_roles (user_id,role_id)(?,(select id from roles where code='cloud_user'))"), lId).Exec()
		}
	}
	return

}

func LdapCreateUser(ldapId int64, userName string, password string, description string) error {

	o := orm.NewOrm()
	o.Using("default")
	admin_username, admin_password, person_dn, host := "", "", "", ""
	err := o.Raw(DbBindReplace("select admin_username,admin_password,person_dn,host from ldaps where id=?"), ldapId).QueryRow(&admin_username, &admin_password, &person_dn, &host)
	if err != nil {
		return err
	}
	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}
	err = l.Bind(admin_username, admin_password)
	if err != nil {
		return err
	}
	control := []ldap.Control{}
	add := ldap.NewAddRequest("cn="+userName+","+person_dn, control)
	//add.Attributes = append(add.Attributes,ldap.Attribute{"description",[]string {description} } )
	add.Attributes = append(add.Attributes, ldap.Attribute{"uid", []string{userName}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"uidNumber", []string{"48890"}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"gidNumber", []string{"0"}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"homeDirectory", []string{"/home/" + userName}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"cn", []string{userName}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"sn", []string{userName}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"objectClass", []string{"inetOrgPerson", "posixAccount", "top"}})
	add.Attributes = append(add.Attributes, ldap.Attribute{"userPassword", []string{password}})
	err = l.Add(add)
	return err
}

func LdapUserForceResetPassword(ldapId int64, userName string, password string) error {

	o := orm.NewOrm()
	o.Using("default")
	admin_username, admin_password, person_dn, host := "", "", "", ""
	err := o.Raw(DbBindReplace("select admin_username,admin_password,person_dn,host from ldaps where id=?"), ldapId).QueryRow(&admin_username, &admin_password, &person_dn, &host)
	if err != nil {
		return err
	}
	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}
	err = l.Bind(admin_username, admin_password)
	if err != nil {
		return err
	}
	control := []ldap.Control{}
	mod := ldap.NewModifyRequest("cn="+userName+","+person_dn, control)
	mod.Replace("userPassword", []string{password})
	err = l.Modify(mod)
	return err
}

func LdapUserChangeMyPassword(ldapId int64, userName string, oldPassword, NewPassword string) error {

	o := orm.NewOrm()
	o.Using("default")
	person_dn, host := "", ""
	err := o.Raw(DbBindReplace("select person_dn,host from ldaps where id=?"), ldapId).QueryRow(&person_dn, &host)
	if err != nil {
		return err
	}
	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}
	err = l.Bind("cn="+userName+","+person_dn, oldPassword)
	if err != nil {
		return err
	}
	control := []ldap.Control{}
	mod := ldap.NewModifyRequest("cn="+userName+","+person_dn, control)
	mod.Replace("userPassword", []string{NewPassword})
	err = l.Modify(mod)
	return err
}

type TLdapUser struct {
	CN                  string
	Manager             string
	Title               string
	Company             string
	Department          string
	DN                  string
	Mail                string
	Description         string
	TelephoneNumber     string
	HomePage            string
	DisplayName         string
	SamAccountName      string
	ExtensionAttribute1 string
	ExtensionAttribute2 string
	ExtensionAttribute3 string
	ExtensionAttribute4 string
}

type TLdapGroup struct {
	CN             string
	DN             string
	Description    string
	SamAccountName string
}

func GetAllGroupsFromAD(l *ldap.Conn, baseDN string) ([]TLdapGroup, error) {
	attributes := []string{"CN", "DESCRIPTION", "SAMACCOUNTNAME"}
	filter := "(objectclass=group)"
	search := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil)
	sr, err := l.Search(search)
	if err != nil {
		return nil, err
	}
	var res []TLdapGroup

	for _, v := range sr.Entries {

		res = append(res, TLdapGroup{
			CN:             v.GetAttributeValue("cn"),
			DN:             v.DN,
			Description:    v.GetAttributeValue("description"),
			SamAccountName: v.GetAttributeValue("samAccountName"),
		})

	}
	return res, nil
}

func GetAllUsersFromAD(l *ldap.Conn, baseDN string) ([]TLdapUser, error) {
	var res []TLdapUser
	attributes := []string{"CN", "MANAGER", "TITLE", "COMPANY", "DEPARTMENT", "MAIL", "DESCRIPTION", "TELEPHONENUMBER", "WWWHOMEPAGE", "DISPLAYNAME", "SAMACCOUNTNAME", "EXTENSIONATTRIBUTE1", "EXTENSIONATTRIBUTE2", "EXTENSIONATTRIBUTE3", "EXTENSIONATTRIBUTE4"}
	//filter := fmt.Sprintf("(&(objectClass=organizationalPerson)(userAccountControl=512))")
	//filter := GetParamValue("ldap_filter")
	//filter := "(objectClass=members)"
	filter := fmt.Sprintf("(objectClass=organizationalPerson)")

	search := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, math.MaxInt32, 0, false,
		filter,
		attributes,
		nil)
	sr, err := l.SearchWithPaging(search, 50)

	if err != nil {
		log.Println("GetAllUsersFromAD error ", err)
		return nil, err
	}
	log.Println("GetAllUsersFromAD ok ")
	for _, v := range sr.Entries {
		res = append(res, TLdapUser{
			CN:                  v.GetAttributeValue("cn"),
			DN:                  v.DN,
			Manager:             v.GetAttributeValue("manager"),
			Title:               v.GetAttributeValue("title"),
			Company:             v.GetAttributeValue("company"),
			Department:          v.GetAttributeValue("department"),
			Mail:                v.GetAttributeValue("mail"),
			Description:         v.GetAttributeValue("description"),
			TelephoneNumber:     v.GetAttributeValue("telephoneNumber"),
			HomePage:            v.GetAttributeValue("wWWHomePage"),
			DisplayName:         v.GetAttributeValue("displayName"),
			SamAccountName:      v.GetAttributeValue("sAMAccountName"),
			ExtensionAttribute1: v.GetAttributeValue("extensionAttribute1"),
			ExtensionAttribute2: v.GetAttributeValue("extensionAttribute2"),
			ExtensionAttribute3: v.GetAttributeValue("extensionAttribute3"),
			ExtensionAttribute4: v.GetAttributeValue("extensionAttribute4"),
		})

	}
	return res, nil

}

func GetAllUsersFromADByGroup(l *ldap.Conn, baseDN, groupDN string) ([]TLdapUser, error) {
	var res []TLdapUser
	attributes := []string{"CN", "MANAGER", "TITLE", "COMPANY", "DEPARTMENT", "MAIL", "DESCRIPTION", "TELEPHONENUMBER", "WWWHOMEPAGE", "DISPLAYNAME", "SAMACCOUNTNAME"}
	filter := fmt.Sprintf("(&(objectClass=person)(memberOf=" + groupDN + "))")
	//filter = "(objectClass=group)"
	search := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil)
	sr, err := l.Search(search)

	if err != nil {
		return nil, err
	}
	for _, v := range sr.Entries {
		res = append(res, TLdapUser{
			CN:              v.GetAttributeValue("cn"),
			DN:              v.DN,
			Manager:         v.GetAttributeValue("manager"),
			Title:           v.GetAttributeValue("title"),
			Company:         v.GetAttributeValue("company"),
			Department:      v.GetAttributeValue("department"),
			Mail:            v.GetAttributeValue("mail"),
			Description:     v.GetAttributeValue("description"),
			TelephoneNumber: v.GetAttributeValue("telephoneNumber"),
			HomePage:        v.GetAttributeValue("wWWHomePage"),
			DisplayName:     v.GetAttributeValue("displayName"),
			SamAccountName:  v.GetAttributeValue("sAMAccountName"),
		})

	}
	return res, nil

}

func LdapSyncAd(ldapUUID string) error {

	err := LdapSyncPersonsAd(ldapUUID)
	if err != nil {
		return err
	}

	err = LdapSync2Users(ldapUUID)
	if err != nil {
		return err
	}

	return nil

	err = LdapSyncGroupsAd(ldapUUID)
	if err != nil {
		return err
	}
	err = LdapSyncPersonsMembersAd(ldapUUID)
	if err != nil {
		return err
	}

	err = LdapSync2Roles(ldapUUID)
	if err != nil {
		return err
	}

	return nil
}

func LdapSyncPersonsAd(ldapUUID string) error {

	log.Println("LdapSyncPersonsAd")
	o := orm.NewOrm()
	o.Using("default")
	ldapId := int64(0)
	host, personDn, adminuserName, adminPassword, pre := "", "", "", "", ""
	err := o.Raw(DbBindReplace("select id,host,person_dn,admin_username,admin_password,pre from ldaps where sys$uuid=?"), ldapUUID).QueryRow(&ldapId, &host, &personDn, &adminuserName, &adminPassword, &pre)

	if err != nil {
		return err
	}

	l, err := ldap.Dial("tcp", host)
	if err != nil {
		log.Println("error!")
		return err
	}
	err = l.Bind(pre+adminuserName, adminPassword)
	if err != nil {
		log.Println("error on Bind " + err.Error())
		return err
	}

	ldapUsers, err := GetAllUsersFromAD(l, personDn)

	if err != nil {
		log.Println("error on GetAllUsersFromAD " + err.Error())
		return err
	}

	log.Println(ldapUsers)

	if len(ldapUsers) > 0 {
		_, err := o.Raw(DbBindReplace("update ldap_persons set is_active=0 where ldap_id=?"), ldapId).Exec()
		if err != nil {
			return err
		}
	}

	for _, v := range ldapUsers {
		_, err := o.Raw(DbBindReplace(`insert into ldap_persons
		(is_active,cn,manager,title,company,department,dn,mail,description,telephonenumber,homepage,displayname,samaccountname,ldap_id,eattr1,eattr2,eattr3,eattr4)
		values (1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`), v.CN, v.Manager, v.Title, v.Company, v.Department, v.DN, v.Mail, v.Description, v.TelephoneNumber, v.HomePage, v.DisplayName, v.SamAccountName, ldapId, v.ExtensionAttribute1, v.ExtensionAttribute2, v.ExtensionAttribute3, v.ExtensionAttribute4).Exec()

		if err != nil && IsDuplicateRow(err) {
			_, err := o.Raw(DbBindReplace(`update ldap_persons
		set is_active=1,cn=?,manager=?,title=?,company=?,department=?,mail=?,description=?,telephonenumber=?,homepage=?,displayname=?,samaccountname=?,
		eattr1=?,eattr2=?,eattr3=?,eattr4=?
		where dn=? and ldap_id=?
		`), v.CN, v.Manager, v.Title, v.Company, v.Department, v.Mail, v.Description, v.TelephoneNumber, v.HomePage, v.DisplayName, v.SamAccountName,
				v.ExtensionAttribute1, v.ExtensionAttribute2, v.ExtensionAttribute3, v.ExtensionAttribute4,
				v.DN, ldapId).Exec()

			if err != nil {
				return err
			}

		} else if err != nil {
			return err
		}
	}

	_, err = o.Raw(DbBindReplace("delete from ldap_persons where samaccountname like '%$' and ldap_id=?"), ldapId).Exec()
	if err != nil {
		return err
	}

	_, err = o.Raw(DbBindReplace(`update  ldap_persons a  JOIN (select id,(select id from ldap_persons zzz where zzz.dn=zz.manager and ldap_id=?) manager_id from ldap_persons zz where ldap_id=?)  z
	on z.id=a.id
	set a.manager_id=z.manager_id
	where ldap_id=?`), ldapId, ldapId, ldapId).Exec()

	if err != nil {
		return err
	}
	return nil

}

func LdapSyncPersonsMembersAd(ldapUUID string) error {

	log.Println("LdapSyncPersonsMembersAd")
	o := orm.NewOrm()
	o.Using("default")
	ldapId := int64(0)
	host, personDn, adminuserName, adminPassword, pre := "", "", "", "", ""
	err := o.Raw(DbBindReplace("select id,host,person_dn,admin_username,admin_password,pre from ldaps where sys$uuid=?"), ldapUUID).QueryRow(&ldapId, &host, &personDn, &adminuserName, &adminPassword, &pre)

	if err != nil {
		return err
	}

	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}
	err = l.Bind(pre+adminuserName, adminPassword)

	if err != nil {
		log.Println("error on Bind " + err.Error())
		return err
	}

	var dns []string
	_, err = o.Raw(DbBindReplace("select dn from ldap_groups where ldap_id=? and is_active=1"), ldapId).QueryRows(&dns)
	if err != nil {
		return err
	}

	for _, groupDn := range dns {
		ldapUsers, err := GetAllUsersFromADByGroup(l, personDn, groupDn)

		if err != nil {
			return err
		}

		log.Println(ldapUsers)

		if len(ldapUsers) > 0 {
			_, err := o.Raw(DbBindReplace("update ldap_person_groups set is_active=0 where group_id in (select id from ldap_groups where dn=? and ldap_id=?) "), groupDn, ldapId).Exec()
			if err != nil {
				return err
			}
		}

		for _, v := range ldapUsers {

			_, err := o.Raw(DbBindReplace(`insert into ldap_person_groups
			(is_active,person_id,group_id)
			values (1,(select id from ldap_persons where dn=?),(select id from ldap_groups where dn=?))
			`), v.DN, groupDn).Exec()

			if err != nil && IsDuplicateRow(err) {

				if v.CN == "test" {
					log.Println("skip")
					log.Println(err)
					log.Println("Title=" + v.CN)
					log.Println("Group=" + groupDn)
				}
				_, err := o.Raw(DbBindReplace(`update ldap_person_groups set is_active=1
				where
				 person_id= (select id from ldap_persons where dn=?) and group_id=(select id from ldap_groups where dn=?)
				`), v.DN, groupDn).Exec()
				if err != nil {
					return err
				}

			} else if err != nil {
				return err

			} else if err != nil {
				return err
			}
		}
	}
	return nil

}

func LdapSyncGroupsAd(ldapUUID string) error {

	log.Println("LdapSyncGroupsAd...")
	o := orm.NewOrm()
	o.Using("default")
	ldapId := int64(0)
	host, personDn, adminuserName, adminPassword, pre := "", "", "", "", ""
	err := o.Raw(DbBindReplace("select id,host,person_dn,admin_username,admin_password,pre from ldaps where sys$uuid=?"), ldapUUID).QueryRow(&ldapId, &host, &personDn, &adminuserName, &adminPassword, &pre)

	if err != nil {
		return err
	}

	l, err := ldap.Dial("tcp", host)
	if err != nil {
		return err
	}
	err = l.Bind(pre+adminuserName, adminPassword)

	if err != nil {
		log.Println("error on Bind " + err.Error())
		return err
	}

	ldapGroups, err := GetAllGroupsFromAD(l, personDn)

	log.Println(ldapGroups)

	if len(ldapGroups) > 0 {
		_, err := o.Raw(DbBindReplace("update ldap_groups set is_active=0 where ldap_id=?"), ldapId).Exec()
		if err != nil {
			return err
		}
	}

	for _, v := range ldapGroups {
		_, err := o.Raw(DbBindReplace(`insert into ldap_groups
		(is_active,cn,dn,description,samaccountname,ldap_id)
		values (1,?,?,?,?,?)
		`), v.CN, v.DN, v.Description, v.SamAccountName, ldapId).Exec()

		if err != nil && IsDuplicateRow(err) {
			_, err := o.Raw(DbBindReplace(`update ldap_groups
		set is_active=1,cn=?,description=?,samaccountname=?
		where dn=? and ldap_id=?
		`), v.CN, v.Description, v.SamAccountName, v.DN, ldapId).Exec()

			if err != nil {
				return err
			}

		} else if err != nil {
			return err
		}
	}
	return nil

}

func LdapSync2Users(ldapUUID string) error {

	log.Println("LdapSync2Users...")
	o := orm.NewOrm()
	o.Using("default")
	_, err := o.Raw(DbBindReplace(`insert into users (login,email,title,mobile,jobtitle,ldap_id)
	(select p.samaccountname,
		coalesce(nullif(p.mail,''),p.samaccountname,p.cn) ,
		coalesce(nullif(p.displayname,''),coalesce(nullif(p.mail,''),p.samaccountname,p.cn),p.samaccountname,p.cn) ,p.telephonenumber,p.title,p.ldap_id
	from ldap_persons p where ldap_id=(select id from ldaps where sys$uuid=?)
	and (select coalesce(p.samaccountname,'') not in (select coalesce(login,'') from users))
	and p.eattr2 in ('S00400','S00086','S00402','S00068','S00052','S00164','S00247','S00305','S00054','S00304','S00232','S00087','S00029','S00010')
	)`), ldapUUID).Exec()

	if err != nil {
		return err
	}
	_, err = o.Raw(DbBindReplace(`update  users a  JOIN
  (select um.id lm_user_id,u.id user_id from  ldap_persons l,ldap_persons m,users um,users u
  where l.manager_id = m.id and um.login=m.samaccountname and u.login=l.samaccountname) z
    on z.user_id=a.id
set a.lm_user_id=z.lm_user_id
where a.ldap_id=(select id from ldaps where sys$uuid=?)`), ldapUUID).Exec()

	return err

}

func LdapSync2Roles(ldapUUID string) error {
	o := orm.NewOrm()
	o.Using("default")

	update_roles := 0
	o.Raw(DbBindReplace("select update_roles from ldaps where sys$uuid=?"), ldapUUID).QueryRow(&update_roles)
	if update_roles != 1 {
		return nil
	}

	log.Println("LdapSync2Roles...")

	_, err := o.Raw(DbBindReplace(`delete from user_roles where user_id in (select id from users where ldap_id=(select id from ldaps where sys$uuid=?))`), ldapUUID).Exec()
	if err != nil {
		return err
	}
	_, err = o.Raw(DbBindReplace(`insert into user_roles
(role_id,user_id,ldap_id)
(
select gr.role_id, (select id from users where login=p.samaccountname) user_id,(select id from ldaps where sys$uuid=?) from ldap_group_roles gr, ldap_person_groups pg,ldap_persons p
where p.ldap_id=(select id from ldaps where sys$uuid=?) and  gr.group_id=pg.group_id and p.id=pg.person_id)`), ldapUUID, ldapUUID).Exec()
	return err

}
