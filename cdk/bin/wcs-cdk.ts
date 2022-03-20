#!/usr/bin/env node
import * as cdk from "@aws-cdk/core";
import { WaterColorSiteStack } from "../lib/wcs-stack";

const targetEnv = process.env.SYSTEM_ENV ? process.env.SYSTEM_ENV : "prod";
const pascalEnv = targetEnv.charAt(0).toUpperCase() + targetEnv.slice(1);

const app = new cdk.App();
new WaterColorSiteStack(app, `WaterColorSiteStack${pascalEnv}`);
