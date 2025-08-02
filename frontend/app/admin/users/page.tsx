"use client";
import React, { useState } from 'react';
import { Search, Filter, MoreHorizontal, User, Mail, Calendar, Shield, Eye, Edit, Trash2, UserCheck, UserX } from 'lucide-react';

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
    <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <User className="w-5 h-5 text-white" />
            </div>
            <div>
              <h2 className="text-2xl font-bold text-gray-900">Users</h2>
              <p className="text-sm text-gray-500">{users.length} total users</p>
            </div>
          </div>
          
          <div className="flex items-center gap-3">
            <button className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition-colors">
              + Add User
            </button>
            <button className="border border-gray-300 px-4 py-2 rounded-lg font-medium hover:bg-gray-50 transition-colors">
              Export
            </button>
          </div>
        </div>
        
        {/* Search and Filters */}
        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <input
              type="text"
              placeholder="Search users..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          
          <button className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors">
            <Filter className="w-4 h-4" />
            <span>Filters</span>
          </button>
        </div>
        
        {/* Bulk Actions */}
        {selectedUsers.length > 0 && (
          <div className="mt-4 p-3 bg-blue-50 rounded-lg flex items-center justify-between">
            <span className="text-sm text-blue-700 font-medium">
              {selectedUsers.length} user{selectedUsers.length > 1 ? 's' : ''} selected
            </span>
            <div className="flex items-center gap-2">
              <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
                Activate
              </button>
              <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
                Deactivate
              </button>
              <button className="text-sm text-red-600 hover:text-red-700 font-medium">
                Delete
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
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
                Apps Created
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
                  <div className="text-sm text-gray-900 font-medium">{user.appsCreated}</div>
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

      {/* Footer */}
      <div className="px-6 py-4 border-t border-gray-200 bg-gray-50">
        <div className="flex items-center justify-between text-sm text-gray-500">
          <span>Showing 1-{users.length} of {users.length} users</span>
          <div className="flex items-center gap-2">
            <button className="px-3 py-1 border border-gray-300 rounded hover:bg-white transition-colors">
              Previous
            </button>
            <button className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors">
              1
            </button>
            <button className="px-3 py-1 border border-gray-300 rounded hover:bg-white transition-colors">
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

