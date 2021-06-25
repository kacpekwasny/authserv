import random
import time
import requests as req

msgFAIL = "fail"

signs=["A","a","B","b","C","c","D","d","E","e","F","f","G","g","H","h","I","i","J","j","K","k","L","l","M","m","N","n","O","o","P","p","R","r","S","s","T","t","U","u","","w","Y","y","Z","z","0","1","2","3","4","5","6","7","8","9","!","@","$","%","^","&","*","(",")","#", ";"]

def genpas(l) -> str:
    paswd = ""
    for _ in range(l):
        time.sleep(.001)
        paswd += random.choice(signs)
    return paswd

def checkFail(dc) -> bool:
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

    def fill(self, dc: dict):
        self.hash_pass = dc["pass_hash"]
        self.token = dc["current_token"]

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

    def cleanDB(self):
        print("Connection.CleanDB()")
        for login in self.getAllLoginsDB():
            self.removeAccount(login)

    def cleanDB_DELETE(self):
        print("DeleteAllRecordsFromDatabase")
        return self.delete("DeleteAllRecordsFromDatabase")

    def getAllLoginsDB(self):
        return self.get("getAllLoginsDB")   

    def getAccount(self, login) -> dict:
        print("getAccount(", login, ")")
        return self.get("getAccount", login=login)
        
    def addAccount(self, login, password) -> dict:
        print("addAccount(", login, password, ")")
        return self.post("addAccount", **{"login":login, "pass":password})

    def removeAccount(self, login):
        print("removeAccount(", login, ")")
        return self.post("removeAccount", login=login)

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

    def get(self, url, **dc_in) -> list:
        dc_in.update(self.credentials)
        ret = req.get(self.url + url, json=dc_in).json()
        if not type(ret) in (dict, list):
            print("Connection.getAllLoginsDB(); ret:", ret)
            return []
        if "status" in ret and ret["status"]==msgFAIL:
            print("Connection.get(); ret:", ret)
            return []
        return ret

    def post(self, url, **dc_in) -> dict:
        dc_in.update(self.credentials)
        return req.post(self.url + url, json=dc_in).json()

    def delete(self, url) -> dict:
        return req.delete(self.url + url, json=self.credentials).json()
    

class Test:
    def __init__(self, url, credentials) -> None:
        self.b = Buff()
        self.c = Connection(url, credentials)
    
    def genAccount(self):
        a = Account()
        self.c.addAccount(a.login, a.password)
        dc = self.c.getAccount(a.login)
        a.fill(dc)
        self.b.appendAcc(a)
    
    def remAccount(self):
        acc = self.randomAcc()
        fail = checkFail( self.c.removeAccount(acc.login) )
        if fail:
            return
        self.b.removeAcc(acc.login)

    def loginAccount(self):
        acc = self.randomAcc()
        dc = self.c.loginAccount(acc.login, acc.password)
        fail = checkFail(dc)
        if fail:
            return
        acc.logged_in = True
        acc.token = dc["token"]

    def prolongAuth(self):
        acc = self.randomAcc()
        dc = self.c.prolongAuth(acc.login, acc.token)
        fail = checkFail(dc)
        if fail:
            if dc["err_code"] == "17":
                if acc.logged_in:
                    print("ERR LOGGED IN BUT UNAUTHENTICATED")
                acc.logged_in = False
                return

    def logoutAccount(self):
        acc = self.randomAcc()
        dc = self.c.logoutAccount(acc.login, acc.token)
        fail = checkFail(dc)
        if fail:
            if dc["err_code"] == "17":
                acc.logged_in = False
                return
        acc.logged_in = False
        
    def changeLogin(self):
        acc = self.randomAcc()
        new_login = Account().login
        dc = self.c.changeLogin(acc.login, acc.token, new_login)
        fail = checkFail(dc)
        if fail:
            print(dc)
            return
        self.b.accs.pop(acc.login)
        acc.login = new_login
        self.b.accs[acc.login] = acc

    def changePassword(self):
        acc = self.randomAcc()
        new_pass = Account().password
        dc = self.c.changePass(acc.login, acc.token, new_pass)
        fail = checkFail(dc)
        if fail:
            print(dc)
            return
        acc.password = new_pass
        dc = self.c.getAccount(acc.login)
        acc.hash_pass = dc["pass_hash"]


    def randomAcc(self) -> Account:
        if len(self.b.accs) > 0:
            return random.choice( list(self.b.accs.values()) )
        return None


    def simulation(self):
        while True:
            if len(self.b.accs)>5:
                func = random.choice([self.remAccount, self.genAccount,  self.loginAccount, self.prolongAuth, self.logoutAccount, 
                                                       self.genAccount, self.loginAccount, self.prolongAuth, 
                                    self.changeLogin, self.changePassword])
                func()
            else:
                self.genAccount()


if __name__ == "__main__":
    t = Test("http://localhost:8888/authserv/", {"client_id": "admin", "client_password": "admin"})
    print(t.c.cleanDB_DELETE())
    print("simulation")
    t.simulation()
