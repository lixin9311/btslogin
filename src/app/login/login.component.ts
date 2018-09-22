import { Challenge, ChallengeResponse } from './challenge';
import { Component, OnInit, Inject } from '@angular/core';
import { Router } from '@angular/router';
import { instantiateSecp256k1, instantiateSha256 } from 'bitcoin-ts';
import { HttpClient, HttpParams } from '@angular/common/http';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material';

export interface DialogData {
  username: string;
  token: string;
}

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})

export class LoginComponent implements OnInit {

  constructor(private router: Router, private http: HttpClient, public dialog: MatDialog) { }
  challengeUrl = 'http://127.0.0.1:8080/challenge';
  username: string;
  password: string;
  message: string;
  token: string;
  // loged = localStorage.getItem('loged') === 'true';
  apikey: string;
  loged = false;

  ngOnInit() { }

  openDialog() {
    const dialogRef = this.dialog.open(LoginDialogComponent, {
      width: '250px',
      data: {username: this.username, token: this.token}
    });
  }

  getUrlParam(key: string): string {
      const temp = decodeURIComponent(window.location.search.substring(1)).split('&')
       .map((v) => v.split('='))
       .filter((v) => (v[0] === key) ? true : false );
      if (temp.length < 1) {
        return '';
      }
      return temp[0][1];
  }

  getChallenge() {
    this.apikey = this.getUrlParam('apikey');
    const params = new HttpParams().set('username', this.username)
      .set('apikey', this.apikey);
      console.log('acuired apikey:', this.apikey);
    return this.http.get(this.challengeUrl, { params: params, responseType: 'text' });
  }

  sendChallenge(signed: string) {
    const challenge: Challenge = {
      apikey: this.apikey,
      username: this.username,
      signed: signed,
    };
    return this.http.post(this.challengeUrl, challenge);
  }

  str2ab(str: string): Uint8Array {
    const buf = new Uint8Array(str.length);
    for (let i = 0, strLen = str.length; i < strLen; i++) {
      buf[i] = str.charCodeAt(i);
    }
    return buf;
  }

  cleanfields() {
    this.message = '';
    this.token = '';
  }

  bs2str(buf: Uint8Array): string {
    return String.fromCharCode.apply(null, buf);
  }

  login(): void {
    this.getChallenge().subscribe(Response => {
      console.log(Response);
      this.message = Response;
      (async () => {
        const secp256k1 = await instantiateSecp256k1();
        const sha256 = await instantiateSha256();
        const privkeyStr = this.username + 'active' + this.password;
        const privkey = sha256.hash(this.str2ab(privkeyStr));
        const hashed = this.str2ab(atob(this.message));
        const signed = secp256k1.signMessageHashDER(privkey, hashed);
        const signedStr = btoa(String.fromCharCode.apply(null, signed));
        console.log(signedStr);
        this.sendChallenge(signedStr).subscribe(
          (data: ChallengeResponse) => {
            console.log(data);
            if (data.code !== 200) {
              alert('wrong credentials');
            } else {
              this.token = data.token;
              this.openDialog();
              // call on success login
            }
          },
          err => {
            alert('login failed');
            this.cleanfields();
            console.log(err);
          }
        );
      })();
    },
      err => {
        alert('login failed');
        this.cleanfields();
        console.log(err);
      });
    // if (this.username === 'admin' && this.password === 'password') {
    //   localStorage.setItem('loged', 'true');
    //   this.router.navigate(['user']);
    // } else {
    //   alert('Invalid credentials!');
    // }
  }
}

@Component({
  selector: 'app-login-dialog',
  templateUrl: './login.dialog.html',
  styleUrls: ['./login.component.css'],
})

export class LoginDialogComponent {
  constructor(
    public dialogRef: MatDialogRef<LoginDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: DialogData) {}
  onNoClick(): void {
    this.dialogRef.close();
  }
}
