import * as cloudfront from "@aws-cdk/aws-cloudfront";
import * as iam from "@aws-cdk/aws-iam";
import * as s3 from "@aws-cdk/aws-s3";
import * as cdk from "@aws-cdk/core";
import * as dynamodb from "@aws-cdk/aws-dynamodb";
import * as acm from "@aws-cdk/aws-certificatemanager";
import * as lambda from "@aws-cdk/aws-lambda";
import * as api from "@aws-cdk/aws-apigateway";
import * as es from '@aws-cdk/aws-elasticsearch';
import { AnyPrincipal } from "@aws-cdk/aws-iam";
import * as logs from '@aws-cdk/aws-logs';

export class WaterColorSiteStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const systemEnv = process.env.SYSTEM_ENV ? process.env.SYSTEM_ENV : "prod";
    const bucketName = `wcs-bucket-${systemEnv}`;
    const wcs = new lambda.Function(this, `wcs-${systemEnv}`, {
      functionName: `wcs-${systemEnv}`,
      runtime: lambda.Runtime.GO_1_X,
      handler: "main",
      code: lambda.Code.fromAsset("../lambda-go/bin"),
      environment: {
        'BUCKET_NAME': bucketName
      },
    });


    const table = new dynamodb.Table(this, `wcs-table-${systemEnv}`, {
      partitionKey: { name: "UserId", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "Timestamp", type: dynamodb.AttributeType.STRING },
      tableName: `wcs-table-${systemEnv}`,
      readCapacity: 1,
      writeCapacity: 1,
    });
    const sessionTable = new dynamodb.Table(this, `wcs-session-table-${systemEnv}`, {
      partitionKey: { name: "SessionId", type: dynamodb.AttributeType.STRING },
      tableName: `wcs-session-table-${systemEnv}`,
      readCapacity: 1,
      writeCapacity: 1,
    });
    const userTable = new dynamodb.Table(this, `wcs-user-table-${systemEnv}`, {
      partitionKey: { name: "UserId", type: dynamodb.AttributeType.STRING },
      tableName: `wcs-user-table-${systemEnv}`,
      readCapacity: 1,
      writeCapacity: 1,
    });

    const northeast1certificate = acm.Certificate.fromCertificateArn(
      this,
      "wcs-certificate-northeast-1",
      "arn:aws:acm:ap-northeast-1:XXXXXXXXXXXX:certificate/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" // *.watercolor.site の証明書。証明書もCDK内部管理にすると作り直すときとかにすごい時間がかかる。
    );
    const esDomain = new es.Domain(this, `wcs-es-domain-${systemEnv}`, {
      version: es.ElasticsearchVersion.V7_10,
      capacity: {
        dataNodes: 1,
        dataNodeInstanceType: 't3.small.elasticsearch'
      },
      domainName: `es-wcs-${systemEnv}`,
      customEndpoint: {
        domainName: 'es.watercolor.site',
        certificate: northeast1certificate,
      },
      accessPolicies: [new iam.PolicyStatement({
        resources: ["arn:aws:es:ap-northeast-1:XXXXXXXXXXXX:domain/es-wcs-prod/*"],
        actions: ['es:*'],
        principals: [new AnyPrincipal()],
        effect: iam.Effect.ALLOW,
        conditions: {
          "IpAddress": { "aws:SourceIp": ["XXX.XXX.XXX.XXX"] } // Lambdaか、オフィスネットワークからならES直アクセス可
        }
      })]
    });

    table.grantFullAccess(wcs);
    sessionTable.grantFullAccess(wcs);
    userTable.grantFullAccess(wcs);
    esDomain.grantReadWrite(wcs);


    const restApiLogAccessLogGroup = new logs.LogGroup(
      this,
      `wcs-api-log-${systemEnv}`,
      {
        logGroupName: `/aws/apigateway/wcs-api-log-${systemEnv}`,
        retention: 1,
      },
    );

    const restapi = new api.RestApi(this, `wcs-api-${systemEnv}`, {
      restApiName: "WaterColorSite API",
      deployOptions: {
        dataTraceEnabled: true,
        loggingLevel: api.MethodLoggingLevel.INFO,
        //アクセスログの設定
        accessLogDestination: new api.LogGroupLogDestination(
          restApiLogAccessLogGroup,
        ),
        accessLogFormat: api.AccessLogFormat.clf(),
      }
    });

    const corsOption = {
      allowOrigins: ["https://watercolor.site"], //静的なサイトのURL。ここからならOK。
      allowHeaders: ["Content-Type"],
      allowMethods: ["POST", "GET", "PATCH"],
      allowCredentials: true,
    }

    const wcsRoot = restapi.root.addResource("wcs");
    wcsRoot.addCorsPreflight(corsOption)
    wcsRoot.addMethod("GET", new api.LambdaIntegration(wcs));
    wcsRoot.addMethod("POST", new api.LambdaIntegration(wcs));

    const getUser = restapi.root.addResource("getUser");
    getUser.addMethod("GET", new api.LambdaIntegration(wcs));

    const paintingOfUser = wcsRoot.addResource("{id}");
    paintingOfUser.addMethod("GET", new api.LambdaIntegration(wcs));
    const aPainting = paintingOfUser.addResource("{timestamp}");
    aPainting.addMethod("GET", new api.LambdaIntegration(wcs));
    const paintingImage = aPainting.addResource("images");
    paintingImage.addCorsPreflight(corsOption)
    paintingImage.addMethod("PATCH", new api.LambdaIntegration(wcs));

    const twitterRoot = restapi.root.addResource("twitter");
    const twitterSignin = twitterRoot.addResource("signin");
    twitterSignin.addMethod("GET", new api.LambdaIntegration(wcs));

    const twitterCallback = twitterRoot.addResource("callback");
    twitterCallback.addMethod("GET", new api.LambdaIntegration(wcs), {
      requestParameters: {
        'method.request.querystring.oauth_token': false,
        'method.request.querystring.oauth_verifier': false,
        'method.request.querystring.denied': false,
      }
    });

    const equipmentRoot = restapi.root.addResource("equipments");
    equipmentRoot.addMethod("GET", new api.LambdaIntegration(wcs), {
      requestParameters: {
        'method.request.querystring.cat': false,
        'method.request.querystring.q': false,
      }
    });


    const bucket = new s3.Bucket(this, `wcs-bucket-${systemEnv}`, {
      bucketName: bucketName,
      versioned: true,
      cors: [{
        allowedMethods: [s3.HttpMethods.POST],
        allowedOrigins: ["*"]
      }]
      // cors: [{
      //   allowedOrigins: ['https://watercolor.site'],
      //   allowedMethods: [s3.HttpMethods.GET, s3.HttpMethods.HEAD]
      // }]
    });

    bucket.grantPut(wcs);

    const oai = new cloudfront.OriginAccessIdentity(
      this,
      `wcs-bucket-oai-${systemEnv}`
    );

    const bucketPolicy = new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ["s3:GetObject"],
      principals: [
        new iam.CanonicalUserPrincipal(
          oai.cloudFrontOriginAccessIdentityS3CanonicalUserId
        ),
      ],
      resources: [bucket.bucketArn + "/*"],
    });
    bucket.addToResourcePolicy(bucketPolicy);

    const certificate = acm.Certificate.fromCertificateArn(
      this,
      "wcs-certificate",
      "arn:aws:acm:us-east-1:XXXXXXXXXXXX:certificate/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" // us-east-1においてある *.watercolor.site の証明書。CloudFrontはリージョンまたがる存在である一方、CertificateManagerは各リージョンにいるので、便宜上us-east-1を決め打ちで参照するようだ。
    );
    new cloudfront.CloudFrontWebDistribution(
      this,
      `wcs-distribution-${systemEnv}`,
      {
        viewerCertificate:
          systemEnv == "prod"
            ? cloudfront.ViewerCertificate.fromAcmCertificate(certificate, {
              aliases: ["img.watercolor.site"],
            })
            : undefined,
        priceClass: cloudfront.PriceClass.PRICE_CLASS_200,
        originConfigs: [
          {
            s3OriginSource: {
              s3BucketSource: bucket,
              originAccessIdentity: oai,
            },
            behaviors: [
              {
                isDefaultBehavior: true,
                minTtl: cdk.Duration.seconds(0),
                maxTtl: cdk.Duration.days(365),
                defaultTtl: cdk.Duration.days(1),
                pathPattern: "*",
                // allowedMethods: cloudfront.CloudFrontAllowedMethods.GET_HEAD_OPTIONS,
                // cachedMethods: cloudfront.CloudFrontAllowedCachedMethods.GET_HEAD_OPTIONS,
                // forwardedValues: {
                //   queryString: false,
                //   headers: [
                //     'Origin',
                //     'Access-Control-Request-Method',
                //     'Access-Control-Request-Headers'
                //   ]
                // }
              },
            ],
          },
        ],
        errorConfigurations: [
          {
            errorCode: 403,
            responsePagePath: "/403.png",
            responseCode: 200,
            errorCachingMinTtl: 0,
          },
          {
            errorCode: 404,
            responsePagePath: "/404.png",
            responseCode: 200,
            errorCachingMinTtl: 0,
          },
        ],
      }
    );
  }
}
