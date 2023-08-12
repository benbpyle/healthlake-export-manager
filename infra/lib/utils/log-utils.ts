import { StageEnvironment } from "../types/stage-environment";

export const getLogLevel = (stage: StageEnvironment): string => {
  switch (stage) {
    case StageEnvironment.DEV:
    case StageEnvironment.QA:
      return "debug";
  }

  return "error";
};
