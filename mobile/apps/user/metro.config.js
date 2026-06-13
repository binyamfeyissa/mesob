const { getDefaultConfig } = require("expo/metro-config");
const path = require("path");

const workspaceRoot = path.resolve(__dirname, "../..");
const projectRoot = __dirname;

// expo-router needs a static app root; set before getDefaultConfig is called
process.env.EXPO_ROUTER_APP_ROOT = "./app";

const config = getDefaultConfig(projectRoot);

// Let Metro see files outside apps/user (shared src/, etc.)
config.watchFolders = [workspaceRoot];

config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, "node_modules"),
  path.resolve(workspaceRoot, "node_modules"),
];

// expo-router 4.x imports @expo/metro-runtime/symbolicate which was removed
// as a root export in @expo/metro-runtime 5.x. Redirect it to the source file.
const _originalResolveRequest = config.resolver.resolveRequest;
config.resolver.resolveRequest = (context, moduleName, platform) => {
  if (moduleName === "@expo/metro-runtime/symbolicate") {
    return {
      filePath: path.resolve(workspaceRoot, "node_modules/@expo/metro-runtime/src/symbolicate.ts"),
      type: "sourceFile",
    };
  }
  if (_originalResolveRequest) {
    return _originalResolveRequest(context, moduleName, platform);
  }
  return context.resolveRequest(context, moduleName, platform);
};

module.exports = config;
