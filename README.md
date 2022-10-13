# CostPilot

Language
----

English | [中文](https://github.com/galaxy-future/costpilot/blob/dev/docs/CH-README.md)

Introduction
-----
CostPilot is an all-in-one cloud cost management platform that applies to applications hosted on all cloud platforms. Developed by Galaxy-Future based on the FinOps discipline, CostPilot helps enterprises cut cloud spending by up to 50%. This platform provides deep insights into your cloud cost, periodic reports that contain both cloud cost data and custom cost optimization suggestions from multiple perspectives, and even state-of-the-art techniques that you can use to efficiently optimize cloud usage.

Getting Started Guide
----
#### 1. Configuration Requirements

To ensure stable operation of the system, we recommend that you use a server with at least 2 CPU cores and 4GB RAM. For Linux and MacOS systems, CostPilot has already been tested.


#### 2. Run in Source Code
To run CostPilot in source code, you need to install Git ( [Git - Downloads](https://git-scm.com/downloads)) and Go (**version 1.17+ is required**) and set your Go workspace first.

* (1) Download source code
  > git clone https://github.com/galaxy-future/costpilot.git
* (2) Use environment variables. You can analyze only one cloud account at a time if you use this method.
  - Replace 'abc' with your own AK/SK.
  ```shell
       COSTPILOT_PROVIDER=AlibabaCloud COSTPILOT_AK=abc COSTPILOT_SK=abc COSTPILOT_REGION_ID=cn-beijing go run .
  ```
* (3) Use a configuration file. You can analyze multiple cloud accounts at a time if you use this method.
    -  Edit conf/config.yaml as follows. You can add multiple items in cloud_accounts. Only AlibabaCloud is supported as provider for now.
     ```yaml
        cloud_accounts:
          - provider:  # required :AlibabaCloud
            ak:  # required
            sk:  # required
            region_id:  # required
            name:  # not required
    ```
    - Execute the following make command:
        ```shell
        make build && make run
      ```
    - After the analysis is complete, visit website/index.html in a browser to view the analysis result.

#### 3. Run in Docker
To run CostPilot in Docker, you need to install Docker first. For more information, see
[Docker Engine Install](https://docs.docker.com/engine/install/).

* (1) Use environment variables. You can analyze only one cloud account at a time if you use this method.
  - Replace 'abc' with your own AK/SK.
    ```shell
    docker run --env COSTPILOT_PROVIDER=AlibabaCloud --env COSTPILOT_AK=abc --env COSTPILOT_SK=abc --env COSTPILOT_REGION_ID=cn-beijing -p 8504:8504 --name=costpilot galaxy-future/costpilot
    ```
* (2) Use a configuration file. You can analyze multiple cloud accounts at a time if you use this method.
  - Create your own config.yaml file, and then execute the following command. Replace /tmp/config.yaml with the absolute path of your config.yaml file.
    ```shell
    docker run --mount type=bind,source=/tmp/config.yaml,target=/home/tiger/app/conf/config.yaml -p 8504:8504 --name=costpilot galaxy-future/costpilot
    ```
#### 4. Sample Result
![costanalysis](https://user-images.githubusercontent.com/87166597/195545060-a341307d-6dca-45f9-8f26-38a484073f5d.jpg)

Code of Conduct
------
[Contributor Convention](https://github.com/galaxy-future/costpilot/blob/master/CODE_OF_CONDUCT)

Authorization
-----

CostPilot uses [Apache License 2.0](https://github.com/galaxy-future/costpilot/blob/master/LICENSE) licensing agreement for authorization.

Contact Us
-----

[Weibo](https://weibo.com/galaxyfuture) | [Zhihu](https://www.zhihu.com/org/xing-yi-wei-lai) | [Bilibili](https://space.bilibili.com/2057006251)
| [WeChat Official Account](https://github.com/galaxy-future/comandx/blob/main/docs/resource/wechat_official_account.md)

If you want more information about the service, scan the following QR code to contact us:

![image](https://user-images.githubusercontent.com/102009012/163559389-813afa06-924f-412d-8642-1a0944384f91.png)

