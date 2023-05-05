import moment from "moment";

export const getFormattedDate = (date: string): string => {
    return moment(date).format('DD.MM.YYYY h:mm');
};

export const getDateTimestamp = (date: string): number => {
  return moment(date).unix();
};