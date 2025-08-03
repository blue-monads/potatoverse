"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal, User, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/BigSearchBar';

export default function Page() {
  return (<>
    <UserTable />

  </>)
}


const UserTable = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);

  const users = [
    {
      id: 1,
      name: 'Alex Johnson',
      email: 'alex.johnson@email.com',
      avatar: 'AJ',
      role: 'Admin',
      status: 'Active',
      joinDate: '2024-01-15',
      lastActive: '2 hours ago',
      appsCreated: 12,
      gradient: 'from-blue-500 to-purple-600'
    },
    {
      id: 2,
      name: 'Sarah Chen',
      email: 'sarah.chen@email.com',
      avatar: 'SC',
      role: 'Developer',
      status: 'Active',
      joinDate: '2024-02-20',
      lastActive: '1 day ago',
      appsCreated: 8,
      gradient: 'from-pink-500 to-rose-500'
    },
    {
      id: 3,
      name: 'Mike Rodriguez',
      email: 'mike.rodriguez@email.com',
      avatar: 'MR',
      role: 'User',
      status: 'Inactive',
      joinDate: '2024-01-08',
      lastActive: '1 week ago',
      appsCreated: 3,
      gradient: 'from-green-500 to-teal-500'
    },
    {
      id: 4,
      name: 'Emily Watson',
      email: 'emily.watson@email.com',
      avatar: 'EW',
      role: 'Moderator',
      status: 'Active',
      joinDate: '2024-03-05',
      lastActive: '5 minutes ago',
      appsCreated: 15,
      gradient: 'from-orange-500 to-red-500'
    },
    {
      id: 5,
      name: 'David Kim',
      email: 'david.kim@email.com',
      avatar: 'DK',
      role: 'Developer',
      status: 'Active',
      joinDate: '2024-02-12',
      lastActive: '3 hours ago',
      appsCreated: 6,
      gradient: 'from-indigo-500 to-purple-500'
    },
    {
      id: 6,
      name: 'Lisa Anderson',
      email: 'lisa.anderson@email.com',
      avatar: 'LA',
      role: 'User',
      status: 'Pending',
      joinDate: '2024-03-15',
      lastActive: 'Never',
      appsCreated: 0,
      gradient: 'from-gray-500 to-gray-600'
    }
  ];

  const handleSelectUser = (userId: number) => {
    setSelectedUsers(prev =>
      prev.includes(userId)
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  const handleSelectAll = () => {
    setSelectedUsers(selectedUsers.length === users.length ? [] : users.map(u => u.id));
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'Admin': return 'bg-red-100 text-red-800';
      case 'Moderator': return 'bg-orange-100 text-orange-800';
      case 'Developer': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Active': return 'bg-green-100 text-green-800';
      case 'Inactive': return 'bg-gray-100 text-gray-800';
      case 'Pending': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <>

      <WithAdminBodyLayout
        Icon={User}
        name='Users'
        description="Manage your users, roles, and permissions."
        rightContent={<></>}

      >

        <BigSearchBar
          setSearchText={setSearchTerm}
          searchText={searchTerm}
        />




        {/* Table */}
        <div className="overflow-x-auto max-w-7xl mx-auto w-full">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedUsers.length === users.length}
                    onChange={handleSelectAll}
                    className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                  />
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Active
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {users.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4">
                    <input
                      type="checkbox"
                      checked={selectedUsers.includes(user.id)}
                      onChange={() => handleSelectUser(user.id)}
                      className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 bg-gradient-to-br ${user.gradient} rounded-full flex items-center justify-center text-white font-semibold`}>
                        {user.avatar}
                      </div>
                      <div>
                        <div className="font-semibold text-gray-900">{user.name}</div>
                        <div className="text-sm text-gray-500 flex items-center gap-1">
                          <Mail className="w-3 h-3" />
                          {user.email}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleColor(user.role)}`}>
                      {user.role}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(user.status)}`}>
                      {user.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500 flex items-center gap-1">
                      <Calendar className="w-3 h-3" />
                      {user.lastActive}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button className="p-1 text-gray-400 hover:text-blue-600 transition-colors" title="View">
                        <Eye className="w-4 h-4" />
                      </button>
                      <button className="p-1 text-gray-400 hover:text-green-600 transition-colors" title="Edit">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button className="p-1 text-gray-400 hover:text-red-600 transition-colors" title="Delete">
                        <Trash2 className="w-4 h-4" />
                      </button>
                      <button className="p-1 text-gray-400 hover:text-gray-600 transition-colors" title="More">
                        <MoreHorizontal className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>


      </WithAdminBodyLayout>

    </>
  );
};

