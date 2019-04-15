# Crew on stdin/stdout

Provides optional, very crude persistence via a JSON file.


```
./run.sh
2019-01-18T20:09:36.144206381Z       input {"to":"captain","update":{"c":{"spec":{"inline":<<cat ../../specs/collatz.yaml | yaml2json>>}}}}
2019-01-18T20:09:36.147201666Z       input {"to":"captain","update":{"dc":{"spec":{"inline":<<cat ../../specs/doublecount.yaml | yaml2json>>}}}}
2019-01-18T20:09:36.147962269Z      update c {"SpecSrc":{"inline":{"name":"collatz","doc":"https://en.wikipedia.org...
2019-01-18T20:09:36.160938267Z       input {"to":"captain","update":{"d":{"spec":{"inline":<<cat ../../specs/double.yaml | yaml2json>>}}}}
2019-01-18T20:09:36.161580593Z      update dc {"SpecSrc":{"inline":{"name":"doublecount","doc":"A machine that doubl...
2019-01-18T20:09:36.172865244Z       input {"double":10}
2019-01-18T20:09:36.173278107Z      update d {"SpecSrc":{"inline":{"name":"double","doc":"A machine that double num...
2019-01-18T20:09:36.173406501Z       input {"double":100}
2019-01-18T20:09:36.175959368Z        emit 0,0 {"doubled":20}
2019-01-18T20:09:36.176216236Z        emit 1,0 {"doubled":20}
2019-01-18T20:09:36.176324663Z      update dc {"State":{"node":"listen","bs":{"count":1}}}
2019-01-18T20:09:36.176338506Z      update d {"State":{"node":"start","bs":{}}}
2019-01-18T20:09:36.177022452Z       input {"double":1000}
2019-01-18T20:09:36.177697157Z        emit 0,0 {"doubled":200}
2019-01-18T20:09:36.177713769Z        emit 1,0 {"doubled":200}
2019-01-18T20:09:36.177719201Z      update dc {"State":{"node":"listen","bs":{"count":2}}}
2019-01-18T20:09:36.177724241Z        emit 0,0 {"doubled":2000}
2019-01-18T20:09:36.17772758Z         emit 1,0 {"doubled":2000}
2019-01-18T20:09:36.177731266Z      update dc {"State":{"node":"listen","bs":{"count":3}}}
2019-01-18T20:09:36.177735923Z       input {"collatz":17}
2019-01-18T20:09:36.177741917Z       input {"collatz":5}
2019-01-18T20:09:36.188311423Z        emit 0,0 {"collatz":52}
2019-01-18T20:09:36.188336983Z        emit 1,0 {"collatz":26}
2019-01-18T20:09:36.188342489Z        emit 2,0 {"collatz":13}
2019-01-18T20:09:36.18834623Z         emit 3,0 {"collatz":40}
2019-01-18T20:09:36.188349907Z        emit 4,0 {"collatz":20}
2019-01-18T20:09:36.188353459Z        emit 5,0 {"collatz":10}
2019-01-18T20:09:36.188356994Z        emit 6,0 {"collatz":5}
2019-01-18T20:09:36.188360173Z        emit 7,0 {"collatz":16}
2019-01-18T20:09:36.188363379Z        emit 8,0 {"collatz":8}
2019-01-18T20:09:36.18836662Z         emit 9,0 {"collatz":4}
2019-01-18T20:09:36.188370333Z        emit 10,0 {"collatz":2}
2019-01-18T20:09:36.188373525Z        emit 11,0 {"collatz":1}
2019-01-18T20:09:36.188394591Z      update c {"State":{"node":"start","bs":{}}}
2019-01-18T20:09:36.188400833Z        emit 0,0 {"collatz":16}
2019-01-18T20:09:36.188408299Z        emit 1,0 {"collatz":8}
2019-01-18T20:09:36.188413067Z        emit 2,0 {"collatz":4}
2019-01-18T20:09:36.188418272Z        emit 3,0 {"collatz":2}
2019-01-18T20:09:36.188423512Z        emit 4,0 {"collatz":1}
2019-01-18T20:09:36.188431733Z       input {"to":"timers","makeTimer":{"in":"2s","msg":{"double":10000},"id":"t0"}}
2019-01-18T20:09:36.188462256Z       input {"to":"timers","makeTimer":{"in":"4s","msg":{"collatz":21},"id":"t1"}}
2019-01-18T20:09:36.189217647Z      update timers {"State":{"node":"start","bs":{"timers":{"t0":{"Id":"t0","Msg":{"doubl...
2019-01-18T20:09:36.189245799Z      update timers {"State":{"node":"start","bs":{"timers":{"t0":{"Id":"t0","Msg":{"doubl...
2019-01-18T20:09:36.195110404Z       input {"to":"d","double":3}
2019-01-18T20:09:36.195144402Z       input {"to":"dc","double":4}
2019-01-18T20:09:36.195542312Z        emit 0,0 {"doubled":6}
2019-01-18T20:09:36.195559767Z        emit 0,0 {"doubled":8}
2019-01-18T20:09:36.195565262Z      update dc {"State":{"node":"listen","bs":{"count":4}}}
2019-01-18T20:09:38.213091627Z        emit 0,0 {"doubled":20000}
2019-01-18T20:09:38.213120276Z        emit 1,0 {"doubled":20000}
2019-01-18T20:09:38.213141372Z      update timers {"State":{"node":"","bs":{"timers":{"t1":{"Id":"t1","Msg":{"collatz":2...
2019-01-18T20:09:38.213148878Z      update dc {"State":{"node":"listen","bs":{"count":5}}}
2019-01-18T20:09:40.192858345Z        emit 0,0 {"collatz":64}
2019-01-18T20:09:40.192884791Z        emit 1,0 {"collatz":32}
2019-01-18T20:09:40.192890293Z        emit 2,0 {"collatz":16}
2019-01-18T20:09:40.192894858Z        emit 3,0 {"collatz":8}
2019-01-18T20:09:40.192898685Z        emit 4,0 {"collatz":4}
2019-01-18T20:09:40.192902395Z        emit 5,0 {"collatz":2}
2019-01-18T20:09:40.192906046Z        emit 6,0 {"collatz":1}
2019-01-18T20:09:40.19294322Z       update timers {"State":{"node":"","bs":{"timers":{}}}}

["c",{"node":"start","bs":{}}]
["d",{"node":"start","bs":{}}]
["dc",{"node":"listen","bs":{"count":5}}]
["timers",{"node":"","bs":{"timers":{}}}]
```
