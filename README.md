#CostPilot

Language
----

English | [中文](https://github.com/galaxy-future/costpilot/blob/dev/docs/CH-README.md)

Introduction
-----
CostPilot is an all-in-one cloud cost management platform that applies to applications hosted on all cloud platforms. Developed by Galaxy-Future based on the FinOps discipline, CostPilot helps enterprises cut cloud spending by up to 50%. This platform provides deep insights into your cloud cost, periodic reports that contain both cloud cost data and custom cost optimization suggestions from multiple perspectives, and even state-of-the-art techniques that you can use to efficiently optimize cloud usage.

Getting Started Guide
----
#### 1. Configuration Requirements

For stable operation of the system, the recommended system model is 2 CPU cores and 4G RAM; for Linux and macOS systems, CostPilot has already been tested.


#### 2. Run In Source Code
To run CostPilot in source code, you need to install Git ( [Git - Downloads](https://git-scm.com/downloads)) and Go (**version 1.17+ is required**) and set your Go workspace first.

* (1) Source code download
  > git clone https://github.com/galaxy-future/costpilot.git
* (2) Run With Environmental Configuration (support only one cloud account analysis)
  - replace 'abc' with your own ak/sk
  ```shell
       COSTPILOT_PROVIDER=AlibabaCloud COSTPILOT_AK=abc COSTPILOT_SK=abc COSTPILOT_REGION_ID=cn-beijing go run .
  ```
* (3) Run With File Configuration (support multiple cloud accounts analysis)
    -  edit conf/config.yaml as follow (support provider: AlibabaCloud for now), you can add multiple items in cloud_accounts.
     ```yaml
        cloud_accounts:
          - provider:  # required :AlibabaCloud
            ak:  # required
            sk:  # required
            region_id:  # required
            name:  # not required
    ```
    - execute make command:
        ```shell
        make build && make run
      ```
    - while the analysis completed, visit website/index.html in the browser to see the analysis result.

#### 3. Run In Docker
To run CostPilot in docker, you need to install docker firstly, otherwise, please check
[Docker Engine Install](https://docs.docker.com/engine/install/).

* (1) Run With Environmental Configuration(support only one cloud account analysis)
  - replace 'abc' with your own ak/sk
    ```shell
    docker run --env COSTPILOT_PROVIDER=AlibabaCloud --env COSTPILOT_AK=abc --env COSTPILOT_SK=abc --env COSTPILOT_REGION_ID=cn-beijing -p 8504:8504 --name=costpilot galaxy-future/costpilot
    ```
* (2) Run With File Configuration (support multiple cloud accounts analysis)
  - create your own /tmp/config.yaml, then execute the following command
    ```shell
    docker run --mount type=bind,source=/tmp/config.yaml,target=/home/tiger/app/conf/config.yaml -p 8504:8504 --name=costpilot galaxy-future/costpilot
    ```
#### 4. Result Shows
//todo add

Code of Conduct
------
[Contributor Convention](https://github.com/galaxy-future/costpilot/blob/master/CODE_OF_CONDUCT)

Authorization
-----

CostPilot uses [Apache License 2.0](https://github.com/galaxy-future/costpilot/blob/master/LICENSE) licensing agreement for authorization

Contact us
-----

[Weibo](https://weibo.com/galaxyfuture) | [Zhihu](https://www.zhihu.com/org/xing-yi-wei-lai) | [Bilibili](https://space.bilibili.com/2057006251)
| [WeChat Official Account](https://github.com/galaxy-future/comandx/blob/main/docs/resource/wechat_official_account.md)

If you want more information about the service, scan the following QR code to contact us:

![image](https://user-images.githubusercontent.com/102009012/163559389-813afa06-924f-412d-8642-1a0944384f91.png)

