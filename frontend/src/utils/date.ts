import moment from "moment";
import { getDateTimeFormat, getDateTimePreciseFormat } from "@/localization";

export const getFormattedDate = (date: string): string => {
  return moment(date).format(getDateTimeFormat());
};

export const getFormattedDatePrecise = (date: string): string => {
  return moment(date).format(getDateTimePreciseFormat());
};

export const getDateTimestamp = (date: string): number => {
  return moment(date).unix();
};
