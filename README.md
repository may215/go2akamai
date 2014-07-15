go2akamai
=========

> Allows for easy communication with Akamai's new REST API to purge/invalidate cached objects.

> You can use it as an embed package, or as http request.


How to use:

First you need to change the configuration files to feet your needs:

cdn.conf

    AkamaiUrisPerRequest - How match purge objects request will be sent.
    User - akamai user 
    Password - akamai password
    Type - arl (default) or cpcode
    Action - production (default) or staging
    Domain - remove (default) or invalidate
    RequestTimeOut - request time out in milisecond

http.conf

    ServerPort - the port to run the web server
    ProccessCount - how match proccesses(cores) to use.
    
Http flush example

    POST http://localhost:8181/flush -d {"type":"arl","action":"invalidate","domain":"production","objects":["http://www.example.com/file1.png","http://www.example.com/file2.png"]}
    
Http flush response
    
    {
        "httpStatus": 201,
        "detail": "Request accepted.",
        "estimatedSeconds": 420,
        "purgeId": "29f21b05-0c1b-11e4-b514-3e7219046db3",
        "progressUri": "/ccu/v2/purges/29f21b05-0c1b-11e4-b514-3e7219046db3",
        "pingAfterSeconds": 420,
        "supportId": "17PY1405427152741501-373851232"
    }
    
Http flush status example

    ** http://localhost:8181/flush_status?stat_id={progressUri}
    GET http://localhost:8181/flush_status?stat_id=/ccu/v2/purges/29f21b05-0c1b-11e4-b514-3e7219046db3

Http flush response

    {
        "originalEstimatedSeconds": 420,
        "progressUri": "",
        "originalQueueLength": 0,
        "purgeId": "fa6a8bc9-0c1c-11e4-9e8c-4a6a5ad5a345",
        "supportId": "17SY1405435673988297-365782112",
        "httpStatus": 200,
        "completionTime": "2014-07-15T12:41:19Z",
        "submittedBy": "your_user",
        "purgeStatus": "Done",
        "submissionTime": "2014-07-15T12:38:52Z",
        "pingAfterSeconds": 0
    }
    
Http queue status example

    GET http://localhost:8181/flush_queue_status

Http queue status response
    
    {
        "httpStatus": 200,
        "queueLength": 0,
        "detail": "The queue may take a minute to reflect new or removed requests.",
        "supportId": "17QY1405427499956195-357389408"
    }

Version
----

1.0

License
----

MIT

Author
----

Meir Shamay @meir_shamay

**Free Software, Hell Yeah!**

[@meir_shamay]:https://www.twitter.com/meir_shamay