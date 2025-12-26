type Props = {
  handleTabClick: (tab: "buy" | "sell") => void;
  selectedTab: "buy" | "sell";
};

export const TabsTransaction = ({ handleTabClick, selectedTab }: Props) => {
  return (
    <div className="flex">
      <button
        onClick={() => handleTabClick("buy")}
        className={`bg-muted cursor-pointer py-1.5 px-3 rounded-l-2xl`}
      >
        <span
          className={`text-sm py-1 px-2 rounded-md ${
            selectedTab === "buy" ? "bg-primary text-white" : ""
          }`}
        >
          Buy
        </span>
      </button>
      <button
        onClick={() => handleTabClick("sell")}
        className={`bg-muted cursor-pointer py-1.5 px-3 rounded-r-2xl`}
      >
        <span
          className={`text-sm py-1 px-2 rounded-md ${
            selectedTab === "sell" ? "bg-primary text-white" : ""
          }`}
        >
          Sell
        </span>
      </button>
    </div>
  );
};
