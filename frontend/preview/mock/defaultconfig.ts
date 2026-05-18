// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import backgroundsJson from "../../../pkg/dconfig/defaultconfig/backgrounds.json";
import mimetypesJson from "../../../pkg/dconfig/defaultconfig/mimetypes.json";
import presetsJson from "../../../pkg/dconfig/defaultconfig/presets.json";
import settingsJson from "../../../pkg/dconfig/defaultconfig/settings.json";
import termthemesJson from "../../../pkg/dconfig/defaultconfig/termthemes.json";
import widgetsJson from "../../../pkg/dconfig/defaultconfig/widgets.json";

export const DefaultFullConfig: FullConfigType = {
    settings: settingsJson as SettingsType,
    mimetypes: mimetypesJson as unknown as { [key: string]: MimeTypeConfigType },
    defaultwidgets: widgetsJson as unknown as { [key: string]: WidgetConfigType },
    widgets: {},
    presets: presetsJson as unknown as { [key: string]: MetaType },
    termthemes: termthemesJson as unknown as { [key: string]: TermThemeType },
    connections: {},
    backgrounds: backgroundsJson as { [key: string]: BackgroundConfigType },
    configerrors: [],
    version: "0.0.0",
    buildtime: "0",
};
