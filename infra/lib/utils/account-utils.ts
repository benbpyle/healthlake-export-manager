import { StageEnvironment } from "../types/stage-environment";

export const getAccountId = (e: StageEnvironment): string => {
  switch (e) {
    case StageEnvironment.DEV:
      return "904442064295";
    case StageEnvironment.QA:
      return "627915793329";
    case StageEnvironment.PROD:
      return "966605421973";
  }

  return "904442064295";
};
