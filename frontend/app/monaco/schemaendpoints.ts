// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import backgroundsSchema from "../../../schema/backgrounds.json";
import settingsSchema from "../../../schema/settings.json";
import widgetsSchema from "../../../schema/widgets.json";

type SchemaInfo = {
    uri: string;
    fileMatch: Array<string>;
    schema: object;
};

const MonacoSchemas: SchemaInfo[] = [
    {
        uri: "dora://schema/settings.json",
        fileMatch: ["*/WAVECONFIGPATH/settings.json"],
        schema: settingsSchema,
    },
    {
        uri: "dora://schema/backgrounds.json",
        fileMatch: ["*/WAVECONFIGPATH/backgrounds.json"],
        schema: backgroundsSchema,
    },
    {
        uri: "dora://schema/widgets.json",
        fileMatch: ["*/WAVECONFIGPATH/widgets.json"],
        schema: widgetsSchema,
    },
];

export { MonacoSchemas };
