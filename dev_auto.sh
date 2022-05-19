#!/usr/bin/expect -f
 set ip 192.168.200.20
 set password ZHU88jie
 set timeout 120
 set user ubuntu

 set path /home/ubuntu/go/process
 set fileName process

 set date [exec date "+%Y_%m_%dT%H:%M:%S"]

proc uploadFile {dst src password} { 
    spawn scp $dst $src
    expect "*password:"
    send "$password\r"
    expect "*100%*"
}

uploadFile  $fileName $user@$ip:$path/$fileName.tmp $password

spawn ssh $user@$ip
expect {
    "*yes/no" { send "yes\r"; exp_continue}
    "*password:" { send "$password\r" }
}
expect "*:*"
send "mv $path/$fileName $path/$fileName.$date\r"
send "mv $path/$fileName.tmp $path/$fileName\r"
send "sudo supervisorctl restart $fileName\r"
expect {
    "*yes/no" { send "yes\r"; exp_continue}
    "*password*" { send "$password\r" }
}
expect {
    "*started" { exp_continue}
    "*file)" { exp_continue}
}
send  "exit\r"
expect eof
