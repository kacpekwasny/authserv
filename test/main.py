import random
import time
import requests as req
from copy import deepcopy

msgFAIL = "fail"

signs=["A","a","B","b","C","c","D","d","E","e","F","f","G","g","H","h","I","i","J","j","K","k","L","l","M","m","N","n","O","o","P","p","R","r","S","s","T","t","U","u","","w","Y","y","Z","z","0","1","2","3","4","5","6","7","8","9","!","@","$","%","^","&","*","(",")","#", ";"]

def genpas(l) -> str:
    paswd = ""
    for _ in range(l):
        time.sleep(.001)
        paswd += random.choice(signs)
    return paswd

def checkFail(dc) -> bool:
    """Return true if fail."""
    if type(dc)==dict:
        if msgFAIL in dc.values():
            if int(dc["err_code"]) in [1,2,3,4,5,7,8,9,10,11,12,13,14,15]:
                print("checkFail(dc); dc:", dc)
            return True
    return False

class Account:
    def __init__(self):
        self.login = genpas(10)
        self.password = genpas(10)
        self.hash_pass = ""
        self.token = ""
        self.logged_in = False
        self.last_login = ""

    def fill(self, dc: dict):
        self.login = dc["login"]
        self.hash_pass = dc["pass_hash"]
        self.token = dc["current_token"]
        self.logged_in = dc["logged_in"]
        self.last_login = dc["last_login"]

class Buff:
    def __init__(self) -> None:
        self.accs={}

    def appendAcc(self, acc):
        self.accs[acc.login] = acc

    def removeAcc(self, login):
        self.accs.pop(login)

    def getAcc(self, login) -> Account:
        return self.accs[login]

class Connection:
    def __init__(self, url, credentials: dict) -> None:
        self.url = url
        self.credentials = credentials

    def cleanDB_DELETE(self):
        print("DeleteAllRecordsFromDatabase")
        return self.delete("DeleteAllRecordsFromDatabase")

    def getAllLoginsDB(self):
        return self.get("getAllLoginsDB")   

    def getAllAccounts(self):
        return self.get("getAllAccountsDB")
        

    def getAccount(self, login):
        print("getAccount(", login, ")")
        return self.get("getAccount", login=login)
        
    def addAccount(self, login, password):
        print("addAccount(", login, password, ")")
        return self.post("addAccount", **{"login":login, "pass":password})

    def removeAccount(self, login):
        print("removeAccount(", login, ")")
        return self.delete("removeAccount", login=login)

    def loginAccount(self, login, password):
        print("loginAccount(", login, password, ")")
        return self.post("loginAccount", **{"login":login, "pass":password})

    def prolongAuth(self, login, token):
        print("prolongAuth(", login, token, ")")
        return self.post("prolongAuth", login=login, token=token)

    def logoutAccount(self, login, token):
        print("logoutAccount(", login, token, ")")
        return self.post("logoutAccount", login=login, token=token)

    def changeLogin(self, login, token, new_login):
        print("changeLogin(", login, "token", new_login, ")")
        return self.post("changeLogin", login=login, token=token, new_login=new_login)

    def changePass(self, login, token, new_pass):
        print("changePass(", login, "token", new_pass, ")")
        return self.post("changePass", login=login, token=token, new_pass=new_pass)

    def addClient(self, client_id, client_password, id_ls):
        return self.post("addClient", add_client_id=client_id, add_client_password=client_password, add_client_access_list=id_ls)

    def removeClient(self, client_id):
        return self.delete("removeClient", remove_client_id=client_id)

    def getClientIDs(self):
        return self.get("getClientIDs")

    def get(self, url, **dc_in) -> (bool and dict):
        dc_in.update(self.credentials)
        ret = req.get(self.url + url, json=dc_in).json()
        return checkFail(ret), ret

    def post(self, url, **dc_in) -> (bool and dict):
        dc_in.update(self.credentials)
        dc = req.post(self.url + url, json=dc_in).json()
        return checkFail(dc), dc

    def delete(self, url, **dc_in) -> (bool and dict):
        dc_in.update(self.credentials)
        dc = req.delete(self.url + url, json=dc_in).json()
        return checkFail(dc), dc

class Test:
    def __init__(self, url, credentials) -> None:
        self.b = Buff()
        self.c = Connection(url, credentials)
    
    def genAccount(self):
        print("\ngenAccount")
        a = Account()
        fail, dc = self.c.addAccount(a.login, a.password)
        if fail:
            print(dc)
            return
        fail, dc = self.c.getAccount(a.login)
        if fail:
            print(dc)
            return
        else:
            dc = dc["additional"]
        a.fill(dc)
        self.b.appendAcc(a)
    
    def remAccount(self):
        print("\nremAccount")
        acc = self.randomAccFailTest()
        fail, dc = self.c.removeAccount(acc.login)
        if fail:
            print(dc)
            return
        self.b.removeAcc(acc.login)

    def loginAccount(self):
        print("\nloginAccount")
        acc = self.randomAccFailTest()
        fail, dc = self.c.loginAccount(acc.login, acc.password)
        if fail:
            print(dc)
            return
        acc.logged_in = True
        acc.token = dc["token"]

    def prolongAuth(self):
        print("\nprolongAuth")
        acc = self.randomAccFailTest()
        fail, dc = self.c.prolongAuth(acc.login, acc.token)
        if fail:
            if dc["err_code"] == "17":
                if acc.logged_in:
                    print("ERR LOGGED IN BUT UNAUTHENTICATED")
                acc.logged_in = False
                return

    def logoutAccount(self):
        print("\nlogoutAccount")
        acc = self.randomAccFailTest()
        dc = self.c.logoutAccount(acc.login, acc.token)
        fail = checkFail(dc)
        if fail:
            if dc["err_code"] == "17":
                acc.logged_in = False
                return
        acc.logged_in = False
        
    def changeLogin(self):
        print("\nchangeLogin")
        acc = self.randomAccFailTest()
        new_login = Account().login
        fail, dc = self.c.changeLogin(acc.login, acc.token, new_login)
        if fail:
            print(dc)
            return
        self.b.accs.pop(acc.login)
        acc.login = new_login
        self.b.accs[acc.login] = acc

    def changePassword(self):
        print("\nchangePassword")
        acc = self.randomAccFailTest()
        new_pass = Account().password
        fail, dc = self.c.changePass(acc.login, acc.token, new_pass)
        if fail:
            print(dc)
            return
        acc.password = new_pass
        fail, dc = self.c.getAccount(acc.login)
        if fail:
            print(dc)
            return
        dc = dc["additional"]
        acc.hash_pass = dc["pass_hash"]


    def syncAccounts(self):
        # Sync is useless because it doesnt know the password, only the hash of it so it cant log in
        fail, dc = self.c.getAllAccounts()
        if fail:
            print(dc)
            return
        dc = dc["additional"]
        print(f"syncAccounts(), type(dc)={type(dc)}, len(dc)={len(dc)}")
        for acc_dc in dc:
            login = acc_dc["login"]
            if login in self.b.accs:
                acc_from_buff = self.b.getAcc(login)
                acc_from_buff.fill(acc_dc)
            else:
                new_acc = Account()
                new_acc.fill(acc_dc)
                self.b.appendAcc(new_acc)

    def randomAcc(self) -> Account:
        if len(self.b.accs) > 0:
            return random.choice( list(self.b.accs.values()) )
        return None

    def randomAccFailTest(self) -> Account:
        """Might return normal account,
        might return a fake account,
        might return an account with altered atributes."""
        i = random.random()
        chance1 = 0.1
        chance2 = 0.1
        if i < chance1:
            print(" ~ ~ Entirely fake account")
            acc = Account()
        elif chance1 < i < chance1 + chance2:
            print(" ~ ~ Partialy fake account")
            acc = deepcopy(self.randomAcc())
            acc.password = "wrong password"
            acc.token = "wrong token"
        else:
            print(" ~ ~ Real account")
            acc = self.randomAcc()
        return acc

    def simulation(self):
        counter = 0
        while True:
            if len(self.b.accs)>5:
                func = random.choice([self.remAccount, self.genAccount,  self.loginAccount, self.prolongAuth, self.logoutAccount, 
                                                       self.genAccount, self.loginAccount, self.prolongAuth, 
                                    self.changeLogin, self.changePassword])
                func()
            else:
                self.genAccount()
            
            if counter>20:
                counter = 0
                print("Buffer size:", len(self.b.accs))
                print("Getting logins from db")
                fail, dc = self.c.getAllLoginsDB()
                if fail:
                    print(dc)
                accs = dc["additional"]
                print("Logins in db:", len(accs))
                time.sleep(2)

            counter += 1


if __name__ == "__main__":
    t = Test("http://localhost:8888/authserv/", {"client_id": "admin", "client_password": "admin"})
    print(t.c.cleanDB_DELETE())
    print(t.c.addClient("kacper", "kacper", list(range(20))))
    print(t.c.getClientIDs())
    t = Test("http://localhost:8888/authserv/", {"client_id": "kacper", "client_password": "kacper"})
    print(t.c.getClientIDs())
    print(t.c.removeClient("admin"))
    print(t.c.getClientIDs())
    print("simulation")
    time.sleep(2)
    t.simulation()
