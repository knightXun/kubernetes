# 0.获取支持创建的卷类型 #
## 请求URL ##
    GET http://dahuacloud.com/v1/volume/type HTTP/1.1
## 请求示例 ##
    GET http://172.25.9.142:56789/v1/volume/type
## 响应示例 ##
    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Tue, 20 Mar 2018 09:59:39 GMT
    Content-Length: 69
    
    {"code":"200","message":"OK","result":{"vol_type":["rbd","dellsc"]}}
## 响应参数 ##
> vol_type：当前支持创建的卷类型，目前支持rbd和dellsc，即ceph开源存储和dell商业存储

----------

# 1.新建卷 #
## 请求URL ##
    POST http://dahuacloud.com/v1/volume HTTP/1.1
## 请求示例 ##
    POST /v1/volume HTTP/1.1
    Content-Type: application/json
    Content-Length: 47
    {"volume":"volume1","size":22,"format":"ext4","uid":"32556","vol_type":"rbd"}
## 请求参数 ##
> volume：卷名称，以大小写字母、数字、横线组合，字母开头，字母或数字结尾
> 
> size：卷大小，单位为G
> 
> format：格式化的文件系统格式，默认为ext4
> 
> uid：创建卷的用户
> 
> vol_type：卷的类型，目前支持rbd，dellsc
## 响应示例 ##
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Sun, 29 Oct 2017 23:09:50 GMT
    
    Content-Length: 23
    
    {"code":"200","message":"OK","result":{"id":"b30c31ad-b8d0-4c63-863e-8cd52a28d9f1"}}

    HTTP/1.1 701 status code 701
    
    Content-Type: application/json
    
    Date: Thu, 09 Nov 2017 19:07:32 GMT
    
    Content-Length: 63
    
    {"code":"701","message":"对应资源已存在","result":null}

----------
# 2.获取卷列表 #
## 请求URL ##
    GET http://dahuacloud.com/v1/volume/list?max=resultnum&pos=position&uid=32556 HTTP/1.1
## 请求示例 ##
    GET /v1/volume/list?max=1&pos=0&uid=32556 HTTP/1.1
## 请求参数 ##
> max：一次获取卷信息的最大条数
> 
> pos：从所有卷列表中获取的起始位置
> 
> uid：创建卷的用户
## 响应示例 ##
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Thu, 09 Nov 2017 19:08:51 GMT
    
    Content-Length: 750
    
    {"code":"200","message":"OK","result":{"total":1,"volumes":[{"attach_status":"detached","create_time":"2018-04-02 17:20:27","owner":"5","provider_misc":null,"size":1,"status":"idle","update_time":"2018-04-02 17:20:27","used_size":0,"vol_type":"dellsc","volid":"8c960b56-2eba-4685-b878-7dcc74082763","volume":"wuchao","volume_mapping":{"access_mode":"ReadWriteOnce","format":"ext4","instance":"","path":""}}]}}
## 响应参数 ##
> total：列表卷记录条数
> 
> volid：卷ID，底层存储的卷名，uuid
> 
> volume：页面显示的卷名
> 
> owner：用户名
> 
> size：卷容量大小
> 
> used_size：卷已使用的容量
> 
> status：卷状态，“busy”代表正被使用(锁定)，“idle”代表未被使用(未锁定)
> 
> attach_status：卷iSCSI映射状态，“attached”代表已被映射，“detached”代表未被映射
> 
> vol_type：卷类型，目前支持rbd(ceph块存储)、dellsc(dell sc系列商业存储)
> 
> provider_misc：卷提供者杂项，包含存储后端需返回给卷使用者需要的信息，不同存储后端、不同卷使用者所需的信息可能不一样。
> 
- rbd返回给k8s的信息包含pool和monitor；示例：

    "provider_misc": {
		"rbd": {
			"image": "563da49b-5095-4411-8441-e1143f7f1a6c",
			"monitors": ["172.25.9.142:6789"],
			"pool": "rbd"
		}
	}
- iSCSI返回的信息包含iscsi，targetPortal，iqn，lun；示例：

        "iscsi": {
    		"locker": "",
    		"target": [{
    			"iqn": "iqn.2002-03.com.compellent:5000d31000d8823e",
    			"lun": 1,
    			"targetPortal": "10.30.72.20:3260"
    		},
    		{
    			"iqn": "iqn.2002-03.com.compellent:5000d31000d8823d",
    			"lun": 1,
    			"targetPortal": "10.30.72.10:3260"
    		},
    		{
    			"iqn": "iqn.2002-03.com.compellent:5000d31000d8823b",
    			"lun": 1,
    			"targetPortal": "10.30.72.10:3260"
    		},
    		{
    			"iqn": "iqn.2002-03.com.compellent:5000d31000d8823c",
    			"lun": 1,
    			"targetPortal": "10.30.72.20:3260"
    		}]
    	}
- FC返回的信息包含fc，targetWWNs，lun；示例：

        "provider_misc": {
    		"fc": {
    			"locker": "",
    			"lun": 1,
    			"targetWWNs": ["5000d31000d88232",
    			"5000d31000d88233",
    			"5000d31000d88231",
    			"5000d31000d88234"]
    		}
    	}
> 
> create_time：创建时间
> 
> update_time：更新时间
> 
> format：文件系统格式
> 
> access_mode：访问模式
> 
> path：卷映射路径
> 
> instance：实例名称，使用卷的服务器

----------
# 3.获取卷信息 #
## 请求URL ##
    GET http://dahuacloud.com/v1/volume/info?volid={volumeid} HTTP/1.1
## 请求示例 ##
    GET http:// dahuacloud.com/v1/volume/info?volid=8c960b56-2eba-4685-b878-7dcc74082763 HTTP/1.1
## 请求参数 ##
> volumeid：卷id
## 响应示例 ##
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Sun, 12 Nov 2017 19:06:17 GMT
    
    Content-Length: 382
    
    {"code":"200","message":"OK","result":{"attach_status":"detached","create_time":"2018-04-02 17:20:27","owner":"5","provider_misc":null,"size":1,"status":"idle","update_time":"2018-04-02 17:20:27","used_size":0,"vol_type":"dellsc","volid":"8c960b56-2eba-4685-b878-7dcc74082763","volume":"wuchao","volume_mapping":{"access_mode":"ReadWriteOnce","format":"ext4","instance":"","path":""}}}

----------
# 4.卷扩容 #
## 请求URL ##
    PUT http://dahuacloud.com/v1/extend/{volumeid}/{size} HTTP/1.1
## 请求示例 ##
    PUT /v1/extend/8a097851-a1bd-4c95-8252-3fadbeba3a0d/35 HTTP/1.1
## 请求参数 ##
> volumeid：卷id
> 
> size：卷扩容后的容量，需大于原容量
## 响应示例 ##
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Wed, 01 Nov 2017 17:35:34 GMT
    
    Content-Length: 44
    
    {"code":"200","message":"OK","result":null}

----------
# 5.删除卷 #
## 请求URL ##
    DELETE http://dahuacloud.com/v1/volume/{volumeid} HTTP/1.1
## 请求示例 ##
    DELETE /v1/volume/8a097851-a1bd-4c95-8252-3fadbeba3a0d HTTP/1.1
## 请求参数 ##
> volumeid：卷id
## 响应示例 ##
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Wed, 01 Nov 2017 17:35:34 GMT
    
    Content-Length: 44
    
    {"code":"200","message":"OK","result":null}

----------
# 6.卷映射(dell存储) #
## 描述 ##
> 目前只支持dell sc商业存储，将dell存储卷通过iSCSI或FC映射至卷使用者所在的服务器,根据服务器支持的hba卡端口类型，优先采用FC的方式，其次iSCSI。K8S在给容器挂载dell商业存储卷之前，需要调用该接口，来获取相应的信息，这些信息包含K8S的FC和iSCSI驱动里yaml配置的必要信息。
## 请求URL ##
    POST http://dahuacloud.com/v1/volume/attach/{volid} HTTP/1.1
## 请求示例 ##
    POST /v1/volume/attach/43eadebe-a2a0-4598-8c08-69fe987bc8ac HTTP/1.1
    
    Content-Type: application/json
    
    Content-Length: 27
    
    {"instance":"10.6.5.205"}
## 请求参数 ##
> volid：卷ID
> 
> instance：卷使用者所在服务器
## 响应示例 ##
### iSCSI映射方式 ###
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Thu, 15 Mar 2018 10:02:21 GMT
    
    Content-Length: 244
    
    {"code":"200","message":"OK","result":{"name":"2e8b94b4-a56a-4cb1-a4c4-b4ca683e8506","iscsi":[{"targetPortal":"10.30.72.10:3260","iqn":"iqn.2002-03.com.compellent:5000d31000d8823d","lun":2},{"targetPortal":"10.30.72.20:3260","iqn":"iqn.2002-03.com.compellent:5000d31000d8823c","lun":2},{"targetPortal":"10.30.72.10:3260","iqn":"iqn.2002-03.com.compellent:5000d31000d8823b","lun":2},{"targetPortal":"10.30.72.20:3260","iqn":"iqn.2002-03.com.compellent:5000d31000d8823e","lun":2}]}}
### FibreChannel映射方式 ###
    HTTP/1.1 200 OK
    
    Content-Type: application/json
    
    Date: Thu, 15 Mar 2018 10:02:21 GMT
    
    Content-Length: 244
    
    {"code":"200","message":"OK","result":{"name":"bbb749b7-9062-4f8a-b518-dc837bc15ef7","fc":{"targetWWNs":["5000D31000D88231","5000D31000D88232","5000D31000D88233","5000D31000D88234"],"lun":2}}}
## 响应参数 ##
### iSCSI映射方式 ###
#### 说明 ####
dell商业存储为了容灾，卷映射后会有多个映射路径，返回的iSCSI映射信息可能包含多组，都是当前连接性较好的映射信息，选用任意一组信息即可。
> name:卷名
> 
> targetPortal:iSCSI target的地址
> 
> iqn:iSCSI target的唯一标识
> 
> lun:卷在iSCSI target的逻辑卷号码
### FibreChannel映射方式 ###
> name：卷名
> 
> targetWWNs：FC target的唯一标识，数组类型，可能包含多个
> 
> lun：卷在FC target的逻辑卷号码

----------

# 7.解除卷映射(dell存储) #
## 描述 ##
> 目前只支持dell sc商业存储，解除卷至使用者所在服务器之间的映射。K8S卸载卷之后，可调用该接口来解除卷的映射；若要删除卷，需先解除卷映射。
## 请求URL ##
    POST http://dahuacloud.com/v1/volume/detach/{volid} HTTP/1.1
## 请求示例 ##
    POST /v1/volume/detach/43eadebe-a2a0-4598-8c08-69fe987bc8ac HTTP/1.1
## 请求参数 ##
> volid：卷ID
## 响应示例 ##
    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Thu, 15 Mar 2018 10:02:21 GMT
    Content-Length: 44

    {"code":"200","message":"OK","result":null}

----------
# 8.锁定卷 #
## 描述 ##
> 目前不允许卷被多点挂载，卷挂载需加锁互斥，以防多点挂载
## 请求URL ##
    POST http://dahuacloud.com/v1/volume/lock HTTP/1.1
## 请求示例 ##
    POST /v1/volume/lock HTTP/1.1
    Content-Type: application/json
    Content-Length: 17
    
    {"id":"2719812c-4e6b-4751-8308-f80140aeecf4","locker":"pod1"}
## 请求参数 ##
> id：卷id
> 
> locker：锁的持有者
## 响应示例 ##
    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Wed, 28 Mar 2018 08:05:07 GMT
    Content-Length: 44
    
    {"code":"200","message":"OK","result":null}
# 9.卷解锁 #
## 请求URL ##
    POST http://dahuacloud.com/v1/volume/unlock HTTP/1.1
## 请求示例 ##
    POST /v1/volume/unlock HTTP/1.1
    Content-Type: application/json
    Content-Length: 17
    
    {"id":"2719812c-4e6b-4751-8308-f80140aeecf4","locker":"pod1"}
## 请求参数 ##
> id：卷id
> 
> locker：锁的持有者
## 响应示例 ##
    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Wed, 28 Mar 2018 08:07:05 GMT
    Content-Length: 57
    
    {"code":"200","message":"OK","result":{"locker":"pod1"}}
## 响应参数 ##
> locker：锁的原持有者
> 