"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal,  UserIcon, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { getUsers, User } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { FantasticTable } from '@/contain';

export default function Page() {
  const [searchTerm, setSearchTerm] = useState('');
  
  const loader  = useSimpleDataLoader<User[]>({
    loader: getUsers,
    ready: true,
  });



 

  return (<>
    <WithAdminBodyLayout
      Icon={UserIcon}
      name='Users'
      description="Manage your users, roles, and permissions."
      rightContent={<>
        <AddButton
          name="+ User"
          onClick={() => { }}
        />

      </>}

    >

      <BigSearchBar
        setSearchText={setSearchTerm}
        searchText={searchTerm}
      />

      <FantasticTable
        isLoading={loader.loading}
        classNamesContainer='max-w-7xl mx-auto w-full py-4 px-2'
        columns={[
          {
            title: 'Name',
            key: 'name',
          },
          {
            title: 'Email',
            key: 'email',
          },
        ]}
        data={loader.data || []}

      />



    </WithAdminBodyLayout>
  </>)
}


