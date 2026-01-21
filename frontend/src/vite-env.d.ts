/// <reference types="vite/client" />

interface Window {
  go: {
    app: {
      App: {
        CheckUpdate: () => Promise<any>;
        DeleteGame: () => Promise<void>;
        DownloadAndLaunch: (username: string, progress: (info: any) => void) => Promise<void>;
        GetCosmeticDatabase: () => Promise<any>;
        GetGamePath: () => Promise<string>;
        GetGameVersion: () => Promise<string>;
        GetSkinPreset: () => Promise<any>;
        IsGameInstalled: () => Promise<boolean>;
        OpenGameFolder: () => Promise<void>;
        RunDiagnostics: () => Promise<any>;
        SaveDiagnostics: () => Promise<string>;
        SaveSkinPreset: (preset: any) => Promise<void>;
        Update: (progress: (info: any) => void) => Promise<void>;
      };
    };
  };
  runtime: {
    Quit: () => void;
    Environment: () => Promise<any>;
    WindowMinimise: () => void;
    WindowMaximise: () => void;
    WindowUnmaximise: () => void;
    WindowToggleMaximise: () => void;
    WindowHide: () => void;
    WindowShow: () => void;
    WindowCenter: () => void;
    WindowSetTitle: (title: string) => void;
    WindowFullscreen: () => void;
    WindowUnfullscreen: () => void;
    WindowIsFullscreen: () => Promise<boolean>;
    WindowIsMaximised: () => Promise<boolean>;
    WindowIsMinimised: () => Promise<boolean>;
    WindowIsNormal: () => Promise<boolean>;
    WindowSetSize: (width: number, height: number) => void;
    WindowGetSize: () => Promise<{w: number; h: number}>;
    WindowSetPosition: (x: number, y: number) => void;
    WindowGetPosition: () => Promise<{x: number; y: number}>;
    WindowSetBackgroundColour: (r: number, g: number, b: number, a: number) => void;
    BrowserOpenURL: (url: string) => void;
    EventsOn: (eventName: string, callback: (...args: any[]) => void) => () => void;
    EventsOnce: (eventName: string, callback: (...args: any[]) => void) => () => void;
    EventsEmit: (eventName: string, ...args: any[]) => void;
    EventsOff: (eventName: string, ...additionalEventNames: string[]) => void;
    ClipboardGetText: () => Promise<string>;
    ClipboardSetText: (text: string) => Promise<boolean>;
    LogPrint: (message: string) => void;
    LogInfo: (message: string) => void;
    LogWarning: (message: string) => void;
    LogError: (message: string) => void;
  };
}
