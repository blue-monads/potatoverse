"use client";

import React, { useState } from "react";
import { Tabs } from "@skeletonlabs/skeleton-react";
import FantasticTable from "@/contain/FantasticTable/FantasticTable";

const Page: React.FC = () => {
  const [group, setGroup] = useState("plane");
  const lorem =
    "Lorem, ipsum dolor sit amet consectetur adipisicing elit. Nostrum veniam reprehenderit eum, reiciendis obcaecati, excepturi nemo ipsa fugit suscipit autem vitae numquam et cumque praesentium vero eos minus itaque. Lorem, ipsum dolor sit amet consectetur adipisicing elit. Nostrum veniam reprehenderit eum, reiciendis obcaecati, excepturi nemo.";

  return (
    // <Tabs value={group} onValueChange={(e) => setGroup(e.value!)}>
    //   <Tabs.List>
    //     <Tabs.Control value="plane">Plane</Tabs.Control>
    //     <Tabs.Control value="boat">Boat</Tabs.Control>
    //     <Tabs.Control value="car">Car</Tabs.Control>
    //   </Tabs.List>
    //   <Tabs.Content>
    //     <Tabs.Panel value="plane">Plane Panel - {lorem}</Tabs.Panel>
    //     <Tabs.Panel value="boat">Boat Panel - {lorem}</Tabs.Panel>
    //     <Tabs.Panel value="car">Car Panel - {lorem}</Tabs.Panel>
    //   </Tabs.Content>
    // </Tabs>

    <>
      <div className="p-4">
        <FantasticTable
          columns={[
            { title: "Position", key: "position" },
            { title: "Symbol", key: "symbol" },
            { title: "Name", key: "name" },
            {
              title: "Atomic No",
              key: "atomic_no",
              render: (data) => <span className="text-right">{data}</span>,
            },
          ]}
          data={[
            { position: 1, symbol: "H", name: "Hydrogen", atomic_no: 1 },
            { position: 2, symbol: "He", name: "Helium", atomic_no: 2 },
            { position: 3, symbol: "Li", name: "Lithium", atomic_no: 3 },
            { position: 4, symbol: "Be", name: "Beryllium", atomic_no: 4 },
            { position: 5, symbol: "B", name: "Boron", atomic_no: 5 },
            { position: 6, symbol: "C", name: "Carbon", atomic_no: 6 },
            { position: 7, symbol: "N", name: "Nitrogen", atomic_no: 7 },
            { position: 8, symbol: "O", name: "Oxygen", atomic_no: 8 },
            { position: 9, symbol: "F", name: "Fluorine", atomic_no: 9 },
            { position: 10, symbol: "Ne", name: "Neon", atomic_no: 10 },
          ]}
          captionText="Periodic Table - First 10 Elements"
          isLoading={false}
          onRowClick={(row) => alert(`You clicked on ${row.name}`)}
          noDataMessage="No elements found."
          actions={[
            {
              label: "Add Element",
              onClick: () => alert("Add Element Clicked"),
              dropdown: false,
            },
            {
              label: "Export",
              onClick: () => alert("Export Clicked"),
              dropdown: true,
            },
          ]}
        />
      </div>
    </>
  );
};

export default Page;
