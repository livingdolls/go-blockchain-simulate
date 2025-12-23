import { FormBuy } from "@/components/organisme/Buy/FormBuy";
import { FormSell } from "@/components/organisme/Sell/FormSell";
import { Card } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function TransactionPage() {
  return (
    <div className="grid grid-cols-12 gap-4">
      <div className="col-span-12 lg:col-span-3">
        <Tabs defaultValue="buy">
          <TabsList>
            <TabsTrigger value="buy">Buy</TabsTrigger>
            <TabsTrigger value="sell">Sell</TabsTrigger>
          </TabsList>
          <TabsContent value="buy">
            <FormBuy />
          </TabsContent>
          <TabsContent value="sell">
            <FormSell />
          </TabsContent>
        </Tabs>
      </div>

      <div className="col-span-12 xl:col-span-9 mt-[45px]">
        <Card className="p-4">
          <h2 className="mb-4 text-lg font-semibold text-center">
            Transaction History
          </h2>
          {/* Transaction history content goes here */}
        </Card>
      </div>
    </div>
  );
}
